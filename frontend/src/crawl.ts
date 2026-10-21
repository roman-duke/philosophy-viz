import type { CrawlMode, WireLink } from './types'

export type CrawlHandlers = {
  onNode?: (key: string, value: WireLink) => void
  onRedirect?: (key: string, address: string) => void
  onLoop?: (key: string) => void
  onError?: (key: string, message: string) => void
  onDone?: () => void
}

type CrawlBody = {
  mode: CrawlMode
  articles: string[]
}

// POST the crawl request and consume the SSE stream via fetch + ReadableStream
// (EventSource can't POST a body). Resolves when the stream ends; rejects on a
// network/HTTP failure. Cancel by aborting the passed signal.
export async function startCrawl(
  body: CrawlBody,
  handlers: CrawlHandlers,
  signal: AbortSignal,
): Promise<void> {
  const res = await fetch('/crawl', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal,
  })

  if (!res.ok || !res.body) {
    throw new Error(`crawl failed (${res.status})`)
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })

    // SSE frames are separated by a blank line.
    let sep: number
    while ((sep = buffer.indexOf('\n\n')) !== -1) {
      const frame = buffer.slice(0, sep)
      buffer = buffer.slice(sep + 2)
      dispatch(frame, handlers)
    }
  }
}

function dispatch(frame: string, h: CrawlHandlers) {
  let event = 'message'
  const dataLines: string[] = []

  for (const line of frame.split('\n')) {
    if (line.startsWith(':')) continue // comment / heartbeat
    if (line.startsWith('event:')) event = line.slice(6).trim()
    else if (line.startsWith('data:')) dataLines.push(line.slice(5).trim())
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let data: any = {}
  const raw = dataLines.join('\n')
  try {
    data = raw ? JSON.parse(raw) : {}
  } catch {
    return // skip malformed frames
  }

  switch (event) {
    case 'node':
      h.onNode?.(data.key, data.link)
      break
    case 'redirect':
      h.onRedirect?.(data.key, data.address)
      break
    case 'loop':
      h.onLoop?.(data.key)
      break
    case 'error':
      h.onError?.(data.key, data.message)
      break
    case 'done':
      h.onDone?.()
      break
  }
}

import { useEffect, useRef, useState } from 'react'
import type {
  Article,
  ChainInfo,
  CrawlMode,
  GraphLinkLite,
  GraphNodeLite,
  NodeKind,
  WireLink,
} from './types'
import { startCrawl } from './crawl'
import { SearchPanel } from './components/SearchPanel'
import { Sidebar } from './components/Sidebar'
import { GraphCanvas } from './components/GraphCanvas'

const MAX = 30
const display = (s: string) => (s ?? '').replace(/_/g, ' ')

type QueuedEvent =
  | { t: 'node'; key: string; value: WireLink }
  | { t: 'redirect'; key: string; address: string }
  | { t: 'loop'; key: string }
  | { t: 'error'; key: string; message: string }

export default function App() {
  const [selected, setSelected] = useState<Article[]>([])
  const [mode, setMode] = useState<CrawlMode>('relaxed')
  const [crawled, setCrawled] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const [nodes, setNodes] = useState<GraphNodeLite[]>([])
  const [links, setLinks] = useState<GraphLinkLite[]>([])
  const [chains, setChains] = useState<Record<string, ChainInfo>>({})
  const [streaming, setStreaming] = useState(false)
  const [streamError, setStreamError] = useState<string | null>(null)

  // authoritative crawl bookkeeping (refs so loop reparenting isn't a setState side-effect)
  const abortRef = useRef<AbortController | null>(null)
  const tailRef = useRef<Record<string, string>>({}) // current main-path tail per seed
  const pathRef = useRef<Record<string, string[]>>({}) // main path per seed

  // SSE hops arrive many-per-second; buffer them and apply on an animation frame
  // so React re-renders at most once per paint instead of once per hop (keeps the
  // main thread free enough that taps — e.g. the sidebar close — aren't dropped).
  const eventBufRef = useRef<QueuedEvent[]>([])
  const rafRef = useRef<number | null>(null)

  useEffect(
    () => () => {
      if (rafRef.current != null) cancelAnimationFrame(rafRef.current)
    },
    [],
  )

  function addArticle(a: Article) {
    setSelected((prev) =>
      prev.length >= MAX || prev.some((p) => p.pageid === a.pageid) ? prev : [...prev, a],
    )
  }
  function removeArticle(pageid: number) {
    setSelected((prev) => prev.filter((p) => p.pageid !== pageid))
  }

  // NewLink: extend the chain's main path; draw edge tail → new node.
  function addHop(key: string, value: WireLink) {
    const title = value?.title
    if (!key || !title) return // ignore malformed node frames
    const prev = tailRef.current[key]
    tailRef.current[key] = title
    pathRef.current[key] = [...(pathRef.current[key] ?? [key]), title]
    const path = pathRef.current[key]

    const kind: NodeKind = title === 'Philosophy' ? 'philosophy' : 'link'
    setNodes((ns) => {
      const found = ns.find((n) => n.id === title)
      if (found) {
        return kind === 'philosophy' && found.kind !== 'philosophy'
          ? ns.map((n) => (n.id === title ? { ...n, kind } : n))
          : ns
      }
      return [
        ...ns,
        {
          id: title,
          label: display(title),
          kind,
          linkAddress: value.linkAddress,
          redirect: value.redirect,
          redirectAddress: value.redirectAddress,
          loopBound: value.loopBound,
        },
      ]
    })
    if (prev) {
      setLinks((ls) =>
        ls.some((l) => l.source === prev && l.target === title)
          ? ls
          : [...ls, { source: prev, target: title }],
      )
    }
    setChains((cs) =>
      cs[key]
        ? { ...cs, [key]: { ...cs[key], path: [...path], done: title === 'Philosophy' || cs[key].done } }
        : cs,
    )
  }

  // Redirect: no new node — just annotate the current tail node.
  function markRedirect(key: string, address: string) {
    if (!key || !address) return
    const tail = tailRef.current[key]
    if (!tail) return
    setNodes((ns) =>
      ns.map((n) => (n.id === tail ? { ...n, redirect: true, redirectAddress: address } : n)),
    )
  }

  // Loop: the tail is a dead-end branch of its parent. Prune it from the main
  // path and reparent the tail so subsequent hops continue from the parent.
  function markLoop(key: string) {
    if (!key) return
    const loopNode = tailRef.current[key]
    const path = pathRef.current[key] ?? [key]
    const parent = path.length >= 2 ? path[path.length - 2] : path[0]
    tailRef.current[key] = parent
    const newPath = path.length >= 2 ? path.slice(0, -1) : path
    pathRef.current[key] = newPath

    if (loopNode) {
      setNodes((ns) =>
        ns.map((n) => (n.id === loopNode ? { ...n, kind: 'loop', loopBound: true } : n)),
      )
    }
    setChains((cs) =>
      cs[key]
        ? {
            ...cs,
            [key]: {
              ...cs[key],
              path: newPath,
              looped: true,
              loops: loopNode ? [...(cs[key].loops ?? []), loopNode] : cs[key].loops,
            },
          }
        : cs,
    )
  }

  function markError(key: string, message: string) {
    setChains((cs) => (cs[key] ? { ...cs, [key]: { ...cs[key], error: message } } : cs))
  }

  // Drain the buffered events in one pass. Running them inside a single tick lets
  // React 18 batch all the setStates into one render.
  function flushEvents() {
    if (rafRef.current != null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
    const batch = eventBufRef.current
    if (batch.length === 0) return
    eventBufRef.current = []
    for (const e of batch) {
      if (e.t === 'node') addHop(e.key, e.value)
      else if (e.t === 'redirect') markRedirect(e.key, e.address)
      else if (e.t === 'loop') markLoop(e.key)
      else markError(e.key, e.message)
    }
  }

  function enqueue(e: QueuedEvent) {
    eventBufRef.current.push(e)
    rafRef.current ??= requestAnimationFrame(flushEvents)
  }

  // drop anything buffered and cancel a pending flush (on abort / new crawl)
  function cancelPending() {
    if (rafRef.current != null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
    eventBufRef.current = []
  }

  function findPaths() {
    if (selected.length === 0) return
    abortRef.current?.abort()
    cancelPending()

    const seeds = selected.map((a) => a.title)
    tailRef.current = Object.fromEntries(seeds.map((s) => [s, s]))
    pathRef.current = Object.fromEntries(seeds.map((s) => [s, [s]]))
    setNodes(seeds.map((s) => ({ id: s, label: display(s), kind: 'seed' as const })))
    setLinks([])
    setChains(Object.fromEntries(seeds.map((s) => [s, { path: [s], done: false, looped: false }])))
    setStreamError(null)
    setStreaming(true)
    setCrawled(true)
    setSidebarOpen(true)

    const ctrl = new AbortController()
    abortRef.current = ctrl
    startCrawl(
      { mode, articles: seeds },
      {
        onNode: (key, value) => enqueue({ t: 'node', key, value }),
        onRedirect: (key, address) => enqueue({ t: 'redirect', key, address }),
        onLoop: (key) => enqueue({ t: 'loop', key }),
        onError: (key, message) => enqueue({ t: 'error', key, message }),
        onDone: () => {
          flushEvents() // apply any buffered tail hops before marking done
          setStreaming(false)
        },
      },
      ctrl.signal,
    ).catch((err) => {
      if (ctrl.signal.aborted) return
      setStreaming(false)
      setStreamError(err instanceof Error ? err.message : String(err))
    })
  }

  function backToSearch() {
    abortRef.current?.abort()
    cancelPending()
    setStreaming(false)
    setCrawled(false)
    setSidebarOpen(false)
  }

  return (
    <div className="h-full p-3 md:p-5">
      <div className="relative mx-auto flex h-full max-w-350 gap-3 md:gap-4">
        {sidebarOpen && (
          <>
            <div
              onClick={() => setSidebarOpen(false)}
              className="fixed inset-0 z-30 bg-ink/30 md:hidden"
              aria-hidden="true"
            />
            <Sidebar
              selected={selected}
              chains={chains}
              onRemove={removeArticle}
              crawled={crawled}
              max={MAX}
              onBackToSearch={backToSearch}
              onClose={() => setSidebarOpen(false)}
            />
          </>
        )}

        <div className="flex min-w-0 flex-1 flex-col gap-3 md:gap-4">
          {!crawled && (
            <SearchPanel
              selected={selected}
              onAdd={addArticle}
              onRemove={removeArticle}
              onFindPaths={findPaths}
              max={MAX}
              mode={mode}
              onModeChange={setMode}
            />
          )}
          <GraphCanvas
            crawled={crawled}
            sidebarOpen={sidebarOpen}
            onToggleSidebar={() => setSidebarOpen((v) => !v)}
            nodes={nodes}
            links={links}
            chains={chains}
            streaming={streaming}
            streamError={streamError}
          />
        </div>
      </div>
    </div>
  )
}

import { useEffect, useRef, useState } from 'react'
import type { Article, SearchResult, TitlesResponse } from '../types'
import { Wordmark } from '../Wordmark'
import { panel } from '../ui'

type Props = {
  selected: Article[]
  onAdd: (a: Article) => void
  onRemove: (pageid: number) => void
  onFindPaths: () => void
  max: number
}

export function SearchPanel({ selected, onAdd, onRemove, onFindPaths, max }: Props) {
  const [keyword, setKeyword] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [justAddedId, setJustAddedId] = useState<number | null>(null)
  const boxRef = useRef<HTMLDivElement>(null)

  const atMax = selected.length >= max
  const selectedIds = new Set(selected.map((s) => s.pageid))
  const query = keyword.trim()
  const showResults = open && query.length >= 2

  // Dismiss the results on an outside click or Escape.
  useEffect(() => {
    function onDocMouseDown(e: MouseEvent) {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false)
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', onDocMouseDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDocMouseDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [])

  // Clear the just-added highlight after it plays once.
  useEffect(() => {
    if (justAddedId == null) return
    const t = setTimeout(() => setJustAddedId(null), 1200)
    return () => clearTimeout(t)
  }, [justAddedId])

  // Search as you type, debounced; abort the in-flight request on each keystroke.
  useEffect(() => {
    if (query.length < 2) {
      setResults([])
      setError(null)
      setLoading(false)
      return
    }
    const ctrl = new AbortController()
    const timer = setTimeout(async () => {
      setLoading(true)
      setError(null)
      try {
        const res = await fetch(`/titles?keyword=${encodeURIComponent(query)}`, {
          signal: ctrl.signal,
        })
        if (!res.ok) throw new Error(String(res.status))
        const data: TitlesResponse = await res.json()
        setResults(data.query?.search ?? [])
      } catch (err) {
        if ((err as Error).name === 'AbortError') return
        setError('Can’t reach the search. Is the Go server running on :8080?')
        setResults([])
      } finally {
        setLoading(false)
      }
    }, 300)
    return () => {
      clearTimeout(timer)
      ctrl.abort()
    }
  }, [query])

  return (
    <section className={`${panel} animate-[fade-rise_0.3s_ease] p-4`}>
      <div className="flex flex-wrap items-center gap-3">
        <Wordmark />
        <div className="ml-auto flex items-center gap-3">
          <span className="font-mono text-xs text-ink-soft">
            {selected.length} / {max}
          </span>
          <button
            onClick={onFindPaths}
            disabled={selected.length === 0}
            className="rounded-full bg-accent px-4 py-1.5 font-sans text-sm font-medium text-paper transition hover:bg-accent-deep focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-40"
          >
            Find paths
          </button>
        </div>
      </div>

      <div ref={boxRef} className="relative mt-3">
        <input
          value={keyword}
          onChange={(e) => {
            setKeyword(e.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          placeholder="Search Wikipedia for an article…"
          className="w-full rounded-xl border border-edge bg-paper-2 px-4 py-2.5 font-sans text-[15px] text-ink outline-none placeholder:text-ink-soft/70 focus:border-accent focus:ring-2 focus:ring-accent/20"
        />

        {showResults && (
          <div className="absolute inset-x-0 top-full z-20 mt-2 max-h-80 overflow-auto rounded-xl border border-edge bg-paper p-1.5 shadow-[0_12px_32px_-12px_rgba(28,24,18,0.45)]">
            {loading && (
              <p className="px-3 py-2 font-mono text-xs text-ink-soft">Searching…</p>
            )}
            {error && <p className="px-3 py-2 text-sm text-danger">{error}</p>}
            {!loading && !error && results.length === 0 && (
              <p className="px-3 py-2 text-sm text-ink-soft">No articles match “{query}”.</p>
            )}
            {!error &&
              results.map((r) => {
                const added = selectedIds.has(r.pageid)
                const blocked = !added && atMax
                return (
                  <button
                    key={r.pageid}
                    onClick={() => {
                      if (added || blocked) return
                      onAdd({ pageid: r.pageid, title: r.title })
                      setJustAddedId(r.pageid) // flash the new pill
                      setKeyword('') // clear the field…
                      setOpen(false) // …and close the dropdown
                    }}
                    disabled={added || blocked}
                    className="flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left transition hover:bg-accent-soft disabled:hover:bg-transparent"
                  >
                    <span className="min-w-0">
                      <span className="block truncate font-display text-[15px] text-ink">
                        {r.title}
                      </span>
                      <span className="font-mono text-[11px] text-ink-soft">
                        {r.wordcount.toLocaleString()} words
                      </span>
                    </span>
                    <span
                      className={`shrink-0 font-mono text-[11px] ${
                        added ? 'text-accent' : blocked ? 'text-ink-soft/60' : 'text-ink-soft'
                      }`}
                    >
                      {added ? 'Added' : blocked ? `Max ${max}` : '+ Add'}
                    </span>
                  </button>
                )
              })}
          </div>
        )}
      </div>

      {selected.length > 0 && (
        <ul className="mt-3 flex flex-wrap gap-2">
          {selected.map((a) => (
            <li key={a.pageid}>
              <button
                onClick={() => onRemove(a.pageid)}
                className={`group flex max-w-64 items-center gap-1.5 rounded-full border border-edge bg-paper-2 py-1 pl-3 pr-2 font-sans text-[13px] text-ink transition hover:border-accent/40 hover:bg-accent-soft ${
                  a.pageid === justAddedId ? 'animate-[pill-glow_1.2s_ease-out]' : ''
                }`}
              >
                <span className="truncate">{a.title}</span>
                <span aria-hidden="true" className="font-mono text-ink-soft transition group-hover:text-accent">
                  ×
                </span>
                <span className="sr-only">Remove {a.title}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

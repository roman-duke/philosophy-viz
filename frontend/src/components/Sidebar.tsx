import { useState } from 'react'
import type { Article } from '../types'
import { Wordmark } from '../Wordmark'
import { panel } from '../ui'

type Props = {
  selected: Article[]
  onRemove: (pageid: number) => void
  crawled: boolean
  max: number
  onBackToSearch: () => void
  onClose: () => void
}

export function Sidebar({ selected, onRemove, crawled, max, onBackToSearch, onClose }: Props) {
  const [openId, setOpenId] = useState<number | null>(null)

  return (
    <aside
      className={`${panel} fixed inset-y-3 left-3 z-40 flex w-72 max-w-[85vw] animate-[slide-in-left_0.3s_ease] flex-col md:static md:inset-auto md:z-auto md:w-80 md:max-w-none`}
    >
      <header className="border-b border-edge p-4">
        <div className="flex items-start justify-between gap-2">
          <Wordmark />
          <button
            onClick={onClose}
            aria-label="Hide selection panel"
            className="shrink-0 rounded-md px-2 py-1 font-mono text-ink-soft transition hover:bg-paper-2 hover:text-ink"
          >
            ×
          </button>
        </div>
        <div className="mt-3 flex items-center justify-between gap-2">
          <p className="font-sans text-[11px] uppercase tracking-[0.14em] text-ink-soft">
            Selection ·{' '}
            <span className="font-mono normal-case tracking-normal text-ink">
              {selected.length}/{max}
            </span>
          </p>
          {crawled && (
            <button
              onClick={onBackToSearch}
              className="rounded-md px-2 py-1 font-sans text-xs font-medium text-accent transition hover:bg-accent-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
              Edit selection
            </button>
          )}
        </div>
      </header>

      <div className="min-h-0 flex-1 overflow-auto p-2">
        {selected.length === 0 ? (
          <p className="px-3 py-6 text-center font-sans text-[13px] leading-relaxed text-ink-soft">
            No articles yet. Add some from the search and they’ll show up here.
          </p>
        ) : (
          selected.map((a) => {
            const isOpen = openId === a.pageid
            return (
              <div key={a.pageid}>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setOpenId(isOpen ? null : a.pageid)}
                    aria-expanded={isOpen}
                    className="flex min-w-0 flex-1 items-center gap-2 rounded-lg px-2.5 py-2 text-left transition hover:bg-paper-2"
                  >
                    <Chevron open={isOpen} />
                    <span className="min-w-0 flex-1 truncate font-display text-[15px] text-ink">
                      {a.title}
                    </span>
                    <span className="shrink-0 font-mono text-[11px] text-ink-soft">
                      {crawled ? '— hops' : ''}
                    </span>
                  </button>
                  <button
                    onClick={() => onRemove(a.pageid)}
                    aria-label={`Remove ${a.title}`}
                    className="shrink-0 rounded-md px-2 py-1 font-mono text-ink-soft transition hover:bg-paper-2 hover:text-ink"
                  >
                    ×
                  </button>
                </div>

                {isOpen && (
                  <div className="animate-[fade-rise_0.2s_ease] pb-3 pl-9 pr-3">
                    <p className="font-sans text-[13px] leading-relaxed text-ink-soft">
                      {crawled
                        ? 'Path tracing isn’t wired up yet — the crawl endpoint is still to come.'
                        : 'Run Find paths to trace this article’s route to Philosophy.'}
                    </p>
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>
    </aside>
  )
}

function Chevron({ open }: { open: boolean }) {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 12 12"
      aria-hidden="true"
      className={`shrink-0 text-ink-soft transition-transform ${open ? 'rotate-90' : ''}`}
    >
      <path
        d="M4 2l4 4-4 4"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

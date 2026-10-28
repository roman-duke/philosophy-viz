import { useState } from 'react'
import type { Article, ChainInfo } from '../types'
import { Wordmark } from '../Wordmark'
import { panel } from '../ui'

const display = (s: string) => s.replace(/_/g, ' ')
const wikiUrl = (title: string) =>
  `https://en.wikipedia.org/wiki/${encodeURIComponent(title.replace(/ /g, '_'))}`

type Props = {
  selected: Article[]
  chains: Record<string, ChainInfo>
  onRemove: (pageid: number) => void
  crawled: boolean
  max: number
  onBackToSearch: () => void
  onClose: () => void
}

export function Sidebar({
  selected,
  chains,
  onRemove,
  crawled,
  max,
  onBackToSearch,
  onClose,
}: Props) {
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
            className="-my-1 -mr-2 flex h-11 w-11 shrink-0 items-center justify-center rounded-md font-mono text-lg leading-none text-ink-soft transition hover:bg-paper-2 hover:text-ink"
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
            const chain = chains[a.title]
            return (
              <div key={a.pageid}>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setOpenId(isOpen ? null : a.pageid)}
                    aria-expanded={isOpen}
                    className="flex min-w-0 flex-1 items-center gap-2 rounded-lg px-2.5 py-2 text-left transition hover:bg-paper-2"
                  >
                    <Chevron open={isOpen} />
                    {crawled && <StatusDot chain={chain} />}
                    <span className="min-w-0 flex-1 truncate font-display text-[15px] text-ink">
                      {a.title}
                    </span>
                    {crawled && chain && (
                      <span className="shrink-0 font-mono text-[11px] text-ink-soft">
                        {Math.max(0, chain.path.length - 1)} hops
                      </span>
                    )}
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
                    {crawled && chain ? (
                      <>
                        <ol className="space-y-1">
                          {chain.path.map((p, i) => (
                            <li key={i} className="flex items-baseline gap-2 font-sans text-[12px]">
                              <span className="w-8 shrink-0 font-mono text-[10px] text-ink-soft/60">
                                {i === 0 ? 'start' : i}
                              </span>
                              <a
                                href={wikiUrl(p)}
                                target="_blank"
                                rel="noopener noreferrer"
                                className={`truncate hover:underline ${
                                  p === 'Philosophy' ? 'text-accent' : 'text-ink hover:text-accent'
                                }`}
                              >
                                {display(p)}
                              </a>
                            </li>
                          ))}
                        </ol>
                        {chain.done && (
                          <p className="mt-2 font-sans text-[12px] text-accent">Reached Philosophy.</p>
                        )}
                        {chain.looped && (
                          <p className="mt-2 font-sans text-[12px] text-[#e6983f]">
                            Hit {chain.loops?.length ?? 1} loop{(chain.loops?.length ?? 1) > 1 ? 's' : ''} along the way.
                          </p>
                        )}
                        {chain.error && (
                          <p className="mt-2 font-sans text-[12px] text-danger">{chain.error}</p>
                        )}
                      </>
                    ) : (
                      <p className="font-sans text-[13px] leading-relaxed text-ink-soft">
                        Run Find paths to trace this article’s route to Philosophy.
                      </p>
                    )}
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

function StatusDot({ chain }: { chain?: ChainInfo }) {
  let color = 'bg-ink-soft/40'
  if (chain?.error) color = 'bg-danger'
  else if (chain?.done) color = 'bg-accent'
  else if (chain?.looped) color = 'bg-[#e6983f]'
  return <span className={`h-1.5 w-1.5 shrink-0 rounded-full ${color}`} aria-hidden="true" />
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

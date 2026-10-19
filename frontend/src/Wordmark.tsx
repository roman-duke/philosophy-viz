// The product mark: a tiny convergence graph — cool outer nodes (one via a
// two-hop chain) feeding into a hot red Philosophy node. Echoes the canvas graph.
export function Wordmark() {
  return (
    <div className="flex items-center gap-2.5">
      <svg width="24" height="24" viewBox="0 0 24 24" aria-hidden="true" className="shrink-0">
        <g stroke="var(--color-ink-soft)" strokeOpacity="0.55" strokeWidth="1" strokeLinecap="round">
          <line x1="4.5" y1="4.5" x2="16" y2="12" />
          <line x1="3.5" y1="19" x2="16" y2="12" />
          <line x1="10.5" y1="21" x2="13.5" y2="15.5" />
          <line x1="13.5" y1="15.5" x2="16" y2="12" />
        </g>
        <g fill="var(--color-ink-soft)">
          <circle cx="4.5" cy="4.5" r="1.7" />
          <circle cx="3.5" cy="19" r="1.7" />
          <circle cx="10.5" cy="21" r="1.7" />
          <circle cx="13.5" cy="15.5" r="1.7" />
        </g>
        <circle cx="16" cy="12" r="3.2" fill="#ff4133" />
      </svg>
      <span className="font-display text-[15px] font-semibold tracking-tight text-ink">
        Path to Philosophy
      </span>
    </div>
  )
}

// The product mark: a tiny convergence graph — cool outer nodes (one via a
// two-hop chain) feeding into a hot red Philosophy node. Echoes the canvas graph.
export function Wordmark() {
  return (
    <div className="flex items-center gap-2.5">
      <svg width="24" height="24" viewBox="0 0 32 32" aria-hidden="true" className="shrink-0">
        <g stroke="var(--color-ink-soft)" strokeOpacity="0.55" strokeWidth="1.9" strokeLinecap="round">
          <line x1="6" y1="15" x2="22" y2="16" />
          <line x1="6" y1="23" x2="22" y2="16" />
          <line x1="7" y1="7" x2="14" y2="9" />
          <line x1="14" y1="9" x2="22" y2="16" />
        </g>
        <g fill="var(--color-ink-soft)">
          <circle cx="6" cy="15" r="2.4" />
          <circle cx="6" cy="23" r="2.4" />
          <circle cx="7" cy="7" r="2.4" />
          <circle cx="14" cy="9" r="2.4" />
        </g>
        <circle cx="22" cy="16" r="4.8" fill="#ff4133" />
      </svg>
      <span className="font-display text-[15px] font-semibold tracking-tight text-ink">
        Path to Philosophy
      </span>
    </div>
  )
}

type Props = {
  onClick: () => void
  open: boolean
}

// Docked control that shows/hides the selection sidebar.
export function SidebarToggle({ onClick, open }: Props) {
  return (
    <button
      onClick={onClick}
      aria-pressed={open}
      aria-label={open ? 'Hide selection panel' : 'Show selection panel'}
      className="shrink-0 rounded-lg border border-edge bg-paper-2 p-2 text-ink-soft transition hover:bg-paper hover:text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
    >
      <svg width="18" height="18" viewBox="0 0 20 20" fill="none" aria-hidden="true">
        <rect x="2.5" y="3.5" width="15" height="13" rx="2.5" stroke="currentColor" strokeWidth="1.4" />
        <line x1="8" y1="3.5" x2="8" y2="16.5" stroke="currentColor" strokeWidth="1.4" />
      </svg>
    </button>
  )
}

import { useMemo } from 'react'
import { panel } from '../ui'
import { SidebarToggle } from './SidebarToggle'

type Props = {
  crawled: boolean
  sidebarOpen: boolean
  onToggleSidebar: () => void
}

type GNode = { id: number; x: number; y: number; layer: number; r: number }
type GLink = { a: number; b: number }

const VIEW_W = 640
const VIEW_H = 460
const CX = VIEW_W / 2
const CY = VIEW_H / 2

// Heat ramp by distance from Philosophy: hot (red) at the core, cooling outward.
const LAYER_COLOR = ['#ff5a3c', '#ff4133', '#f0863f', '#6d8bad'] // 0=core … 3=rim
const LAYER_R = [12, 7, 5.5, 4.5]
const RING_RADIUS = [0, 78, 140, 200]

// Deterministic PRNG so the layout is stable across renders (no jitter on re-render).
function mulberry32(seed: number) {
  return function () {
    seed |= 0
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// A convergence tree: rim → mid → inner → Philosophy, laid out in branching
// clusters so it reads like an Obsidian graph rather than a clean starburst.
function buildGraph(): { nodes: GNode[]; links: GLink[] } {
  const rnd = mulberry32(42)
  const jitter = (amt: number) => (rnd() - 0.5) * 2 * amt

  const nodes: GNode[] = []
  const links: GLink[] = []

  const push = (x: number, y: number, layer: number) => {
    const id = nodes.length
    nodes.push({ id, x, y, layer, r: LAYER_R[layer] })
    return id
  }
  const place = (angle: number, radius: number, layer: number) =>
    push(CX + Math.cos(angle) * radius, CY + Math.sin(angle) * radius, layer)

  const center = push(CX, CY, 0)

  const innerCount = 5
  for (let i = 0; i < innerCount; i++) {
    const a1 = (i / innerCount) * Math.PI * 2 + jitter(0.25)
    const inner = place(a1, RING_RADIUS[1] + jitter(10), 1)
    links.push({ a: center, b: inner })

    const midCount = 2
    for (let j = 0; j < midCount; j++) {
      const a2 = a1 + (j - (midCount - 1) / 2) * 0.42 + jitter(0.12)
      const mid = place(a2, RING_RADIUS[2] + jitter(12), 2)
      links.push({ a: inner, b: mid })

      const outerCount = 1 + Math.round(rnd()) // 1 or 2 leaves per mid node
      for (let k = 0; k < outerCount; k++) {
        const a3 = a2 + (outerCount === 1 ? jitter(0.18) : (k - 0.5) * 0.3 + jitter(0.1))
        const outer = place(a3, RING_RADIUS[3] + jitter(14), 3)
        links.push({ a: mid, b: outer })
      }
    }
  }

  return { nodes, links }
}

export function GraphCanvas({ crawled, sidebarOpen, onToggleSidebar }: Props) {
  const { nodes, links } = useMemo(buildGraph, [])
  // Draw rim nodes first so inner/core nodes overlap cleanly on top.
  const drawOrder = useMemo(() => [...nodes].sort((a, b) => b.layer - a.layer), [nodes])

  return (
    <section className={`${panel} relative flex flex-1 flex-col overflow-hidden`}>
      <div className="absolute left-4 top-4 z-10 flex items-center gap-2.5">
        <SidebarToggle onClick={onToggleSidebar} open={sidebarOpen} />
        <span className="font-sans text-[11px] uppercase tracking-[0.14em] text-ink-soft">
          Canvas
        </span>
      </div>

      {crawled ? (
        <div className="flex flex-1 items-center justify-center">
          <p className="px-8 text-center font-sans text-sm text-ink-soft">
            The crawl runs here. Graph rendering comes next, once the paths endpoint streams in.
          </p>
        </div>
      ) : (
        <>
          <div className="relative min-h-0 flex-1">
            <svg
              viewBox={`0 0 ${VIEW_W} ${VIEW_H}`}
              preserveAspectRatio="xMidYMid meet"
              className="h-full w-full"
              aria-hidden="true"
            >
              <defs>
                <filter id="philosophy-glow" x="-120%" y="-120%" width="340%" height="340%">
                  <feGaussianBlur stdDeviation="8" />
                </filter>
              </defs>

              <g stroke="var(--color-ink-soft)" strokeOpacity="0.22" strokeWidth="1">
                {links.map((l, i) => {
                  const a = nodes[l.a]
                  const b = nodes[l.b]
                  return <line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y} />
                })}
              </g>

              {/* soft pulsing halo behind Philosophy */}
              <circle
                cx={CX}
                cy={CY}
                r={16}
                fill={LAYER_COLOR[0]}
                filter="url(#philosophy-glow)"
                className="animate-[halo-pulse_3.2s_ease-in-out_infinite]"
                style={{ transformBox: 'fill-box', transformOrigin: 'center', opacity: 0.4 }}
              />

              {drawOrder.map((n) => (
                <circle
                  key={n.id}
                  cx={n.x}
                  cy={n.y}
                  r={n.r}
                  fill={LAYER_COLOR[n.layer]}
                  stroke="var(--color-paper)"
                  strokeWidth={n.layer === 0 ? 2 : 1}
                />
              ))}

              <text
                x={CX}
                y={CY + LAYER_R[0] + 20}
                textAnchor="middle"
                className="font-display"
                fontSize="15"
                fontStyle="italic"
                fill="var(--color-ink)"
              >
                Philosophy
              </text>
            </svg>
          </div>

          <div className="shrink-0 px-8 pb-8 text-center">
            <h1 className="font-display text-xl font-semibold tracking-tight text-ink md:text-2xl">
              All roads lead to Philosophy
            </h1>
            <p className="mx-auto mt-2 max-w-md font-sans text-[15px] leading-relaxed text-ink-soft">
              Follow any article’s first link, over and over, and the chain almost always
              ends here. Add articles, then trace their paths.
            </p>
          </div>
        </>
      )}
    </section>
  )
}

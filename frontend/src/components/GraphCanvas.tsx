import { useMemo } from 'react'
import { panel } from '../ui'
import { SidebarToggle } from './SidebarToggle'
import { ForceGraph } from './ForceGraph'
import type { ChainInfo, GraphLinkLite, GraphNodeLite } from '../types'

type Props = {
  crawled: boolean
  sidebarOpen: boolean
  onToggleSidebar: () => void
  nodes: GraphNodeLite[]
  links: GraphLinkLite[]
  chains: Record<string, ChainInfo>
  streaming: boolean
  streamError: string | null
}

// ---- empty-state heat graph (static, decorative) ----
type GNode = { id: number; x: number; y: number; layer: number; r: number }
type GLink = { a: number; b: number }

const VIEW_W = 640
const VIEW_H = 460
const CX = VIEW_W / 2
const CY = VIEW_H / 2
const LAYER_COLOR = ['#ff5a3c', '#ff4133', '#f0863f', '#6d8bad']
const LAYER_R = [12, 7, 5.5, 4.5]
const RING_RADIUS = [0, 78, 140, 200]

function mulberry32(seed: number) {
  return function () {
    seed |= 0
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function buildEmptyGraph(): { nodes: GNode[]; links: GLink[] } {
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
    for (let j = 0; j < 2; j++) {
      const a2 = a1 + (j - 0.5) * 0.42 + jitter(0.12)
      const mid = place(a2, RING_RADIUS[2] + jitter(12), 2)
      links.push({ a: inner, b: mid })
      const outerCount = 1 + Math.round(rnd())
      for (let k = 0; k < outerCount; k++) {
        const a3 = a2 + (outerCount === 1 ? jitter(0.18) : (k - 0.5) * 0.3 + jitter(0.1))
        const outer = place(a3, RING_RADIUS[3] + jitter(14), 3)
        links.push({ a: mid, b: outer })
      }
    }
  }
  return { nodes, links }
}

export function GraphCanvas({
  crawled,
  sidebarOpen,
  onToggleSidebar,
  nodes,
  links,
  chains,
  streaming,
  streamError,
}: Props) {
  const empty = useMemo(buildEmptyGraph, [])
  const drawOrder = useMemo(() => [...empty.nodes].sort((a, b) => b.layer - a.layer), [empty.nodes])

  return (
    <section className={`${panel} relative flex flex-1 flex-col overflow-hidden`}>
      <div className="absolute left-4 top-4 z-10 flex items-center gap-2.5">
        <SidebarToggle onClick={onToggleSidebar} open={sidebarOpen} />
        <span className="font-sans text-[11px] uppercase tracking-[0.14em] text-ink-soft">
          Canvas
          {crawled && (
            <span className="ml-1 normal-case tracking-normal text-ink-soft/70">
              · {streaming ? 'tracing paths…' : 'done'}
            </span>
          )}
        </span>
      </div>

      {streamError && (
        <div className="absolute right-4 top-4 z-10 max-w-xs rounded-lg border border-danger/30 bg-paper px-3 py-1.5 text-[12px] text-danger shadow-sm">
          {streamError}
        </div>
      )}

      {crawled ? (
        <ForceGraph nodes={nodes} links={links} chains={chains} />
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
                {empty.links.map((l, i) => {
                  const a = empty.nodes[l.a]
                  const b = empty.nodes[l.b]
                  return <line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y} />
                })}
              </g>

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

import { useEffect, useRef, useState } from 'react'
import * as d3 from 'd3'
import type { ChainInfo, GraphLinkLite, GraphNodeLite, NodeKind } from '../types'

interface SimNode extends d3.SimulationNodeDatum {
  id: string
  label: string
  kind: NodeKind
  linkAddress?: string
  redirect?: boolean
  redirectAddress?: string
  loopBound?: boolean
}
interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  source: string | SimNode
  target: string | SimNode
}

const RADIUS: Record<NodeKind, number> = { philosophy: 11, seed: 7, loop: 6, link: 5 }
// heat ramp: hot red nearest Philosophy → cool slate at the rim
const heat = d3.interpolateRgbBasis(['#ff4133', '#f0863f', '#6d8bad'])

const fmtAddr = (a?: string) => (a ?? '').replace('/wiki/', '').replace(/_/g, ' ')

type Props = {
  nodes: GraphNodeLite[]
  links: GraphLinkLite[]
  chains: Record<string, ChainInfo>
}

export function ForceGraph({ nodes, links, chains }: Props) {
  const [activeSeed, setActiveSeed] = useState<string | null>(null)
  const wrapRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)
  const rootRef = useRef<SVGGElement>(null)
  const linksRef = useRef<SVGGElement>(null)
  const nodesRef = useRef<SVGGElement>(null)
  const simRef = useRef<d3.Simulation<SimNode, SimLink> | null>(null)
  const nodeMap = useRef<Map<string, SimNode>>(new Map())
  const [hover, setHover] = useState<{ node: SimNode; x: number; y: number } | null>(null)

  // one-time setup: simulation, tick renderer, zoom, resize
  useEffect(() => {
    const svg = d3.select(svgRef.current!)
    const root = d3.select(rootRef.current!)

    const sim = d3
      .forceSimulation<SimNode>()
      .force('charge', d3.forceManyBody().strength(-220))
      .force(
        'link',
        d3
          .forceLink<SimNode, SimLink>()
          .id((d) => d.id)
          .distance(48)
          .strength(0.9),
      )
      .force('collide', d3.forceCollide<SimNode>().radius((d) => RADIUS[d.kind] + 6))
      .force('x', d3.forceX<SimNode>().strength(0.05))
      .force('y', d3.forceY<SimNode>().strength(0.05))
    simRef.current = sim

    sim.on('tick', () => {
      d3.select(linksRef.current!)
        .selectAll<SVGLineElement, SimLink>('line')
        .attr('x1', (d) => (d.source as SimNode).x ?? 0)
        .attr('y1', (d) => (d.source as SimNode).y ?? 0)
        .attr('x2', (d) => (d.target as SimNode).x ?? 0)
        .attr('y2', (d) => (d.target as SimNode).y ?? 0)
      d3.select(nodesRef.current!)
        .selectAll<SVGGElement, SimNode>('g.node')
        .attr('transform', (d) => `translate(${d.x ?? 0},${d.y ?? 0})`)
    })

    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.3, 3])
      .on('zoom', (e: d3.D3ZoomEvent<SVGSVGElement, unknown>) =>
        root.attr('transform', e.transform.toString()),
      )
    svg.call(zoom)
    svg.on('click', () => setActiveSeed(null)) // click empty space clears focus

    const resize = () => {
      if (!wrapRef.current) return // observer can fire during teardown
      const r = wrapRef.current.getBoundingClientRect()
      svg.attr('viewBox', `0 0 ${r.width} ${r.height}`)
      ;(sim.force('x') as d3.ForceX<SimNode>).x(r.width / 2)
      ;(sim.force('y') as d3.ForceY<SimNode>).y(r.height / 2)
      sim.alpha(0.3).restart()
    }
    const ro = new ResizeObserver(resize)
    if (wrapRef.current) ro.observe(wrapRef.current)
    resize()

    return () => {
      ro.disconnect()
      sim.stop()
    }
  }, [])

  // data-driven updates (runs whenever nodes/links change)
  useEffect(() => {
    const sim = simRef.current
    if (!sim) return

    const map = nodeMap.current
    const rect = wrapRef.current?.getBoundingClientRect()
    const cx = (rect?.width ?? 600) / 2
    const cy = (rect?.height ?? 400) / 2

    // reconcile persistent node objects so positions survive re-renders
    for (const n of nodes) {
      const existing = map.get(n.id)
      if (existing) {
        existing.kind = n.kind
        existing.label = n.label
        existing.linkAddress = n.linkAddress
        existing.redirect = n.redirect
        existing.redirectAddress = n.redirectAddress
        existing.loopBound = n.loopBound
      } else {
        map.set(n.id, {
          id: n.id,
          label: n.label,
          kind: n.kind,
          linkAddress: n.linkAddress,
          redirect: n.redirect,
          redirectAddress: n.redirectAddress,
          loopBound: n.loopBound,
          x: cx + (Math.random() - 0.5) * 40,
          y: cy + (Math.random() - 0.5) * 40,
        })
      }
    }

    const simNodes = nodes.map((n) => map.get(n.id)!)
    const idSet = new Set(nodes.map((n) => n.id))
    const validLinks = links.filter((l) => idSet.has(l.source) && idSet.has(l.target))
    const simLinks: SimLink[] = validLinks.map((l) => ({ source: l.source, target: l.target }))

    sim.nodes(simNodes)
    ;(sim.force('link') as d3.ForceLink<SimNode, SimLink>).links(simLinks)

    // colour by hop-distance from Philosophy
    const dist = bfsFromPhilosophy(nodes, validLinks)
    const maxD = Math.max(1, ...Object.values(dist))
    const fill = (n: SimNode) => {
      if (n.kind === 'loop') return '#e6983f'
      const d = dist[n.id]
      return d == null ? '#8aa0bd' : heat(d / maxD)
    }

    d3.select(linksRef.current!)
      .selectAll<SVGLineElement, SimLink>('line')
      .data(simLinks, (l) => `${idOf(l.source)}->${idOf(l.target)}`)
      .join((enter) =>
        enter
          .append('line')
          .attr('stroke', 'var(--color-ink-soft)')
          .attr('stroke-opacity', 0.25)
          .attr('stroke-width', 1),
      )

    const nodeSel = d3
      .select(nodesRef.current!)
      .selectAll<SVGGElement, SimNode>('g.node')
      .data(simNodes, (d) => d.id)
      .join((enter) => {
        const g = enter.append('g').attr('class', 'node').style('cursor', 'grab')
        g.append('circle').attr('stroke', 'var(--color-paper)').attr('stroke-width', 1.5)
        g.append('text')
          .attr('text-anchor', 'middle')
          .attr('font-size', 10)
          .attr('fill', 'var(--color-ink)')
          .style('font-family', 'var(--font-sans)')
          .style('pointer-events', 'none')
        return g
      })

    nodeSel
      .select<SVGCircleElement>('circle')
      .attr('r', (d) => RADIUS[d.kind])
      // loop leaves read as dead-ends: hollow fill + dashed amber border
      .attr('fill', (d) => (d.kind === 'loop' ? 'var(--color-paper)' : fill(d)))
      .attr('stroke', (d) => (d.kind === 'loop' ? '#e6983f' : 'var(--color-paper)'))
      .attr('stroke-dasharray', (d) => (d.kind === 'loop' ? '3 2' : null))
    nodeSel
      .select<SVGTextElement>('text')
      .attr('y', (d) => -(RADIUS[d.kind] + 5))
      .text((d) => (d.kind === 'seed' || d.kind === 'philosophy' ? d.label : ''))

    nodeSel
      .on('mouseenter', (e: MouseEvent, d) => {
        if (!wrapRef.current) return
        const [x, y] = d3.pointer(e, wrapRef.current)
        setHover({ node: d, x, y })
      })
      .on('mousemove', (e: MouseEvent, d) => {
        if (!wrapRef.current) return
        const [x, y] = d3.pointer(e, wrapRef.current)
        setHover({ node: d, x, y })
      })
      .on('mouseleave', () => setHover(null))
      .on('click', (e: MouseEvent, d) => {
        e.stopPropagation() // don't let it bubble to the svg "clear" handler
        if (d.kind === 'seed') setActiveSeed((prev) => (prev === d.id ? null : d.id))
      })
      .call(
        d3
          .drag<SVGGElement, SimNode>()
          .on('start', (e: d3.D3DragEvent<SVGGElement, SimNode, SimNode>, d) => {
            if (!e.active) sim.alphaTarget(0.3).restart()
            d.fx = d.x
            d.fy = d.y
          })
          .on('drag', (e: d3.D3DragEvent<SVGGElement, SimNode, SimNode>, d) => {
            d.fx = e.x
            d.fy = e.y
          })
          .on('end', (e: d3.D3DragEvent<SVGGElement, SimNode, SimNode>, d) => {
            if (!e.active) sim.alphaTarget(0)
            d.fx = null
            d.fy = null
          }),
      )

    sim.alpha(0.7).restart()
  }, [nodes, links])

  // focus mode: clicking a start node keeps its path (+ loop branches) lit, dims the rest
  useEffect(() => {
    const nodesG = nodesRef.current
    const linksG = linksRef.current
    if (!nodesG || !linksG) return

    const chain = activeSeed ? chains[activeSeed] : undefined
    const lit = chain ? new Set<string>([...chain.path, ...(chain.loops ?? [])]) : null

    d3.select(nodesG)
      .selectAll<SVGGElement, SimNode>('g.node')
      .style('opacity', (d) => (!lit || lit.has(d.id) ? 1 : 0.1))
    d3.select(linksG)
      .selectAll<SVGLineElement, SimLink>('line')
      .style('opacity', (d) => {
        if (!lit) return 1
        return lit.has(idOf(d.source)) && lit.has(idOf(d.target)) ? 1 : 0.08
      })
  }, [activeSeed, nodes, links, chains])

  return (
    <div ref={wrapRef} className="absolute inset-0">
      <svg ref={svgRef} className="h-full w-full">
        <g ref={rootRef}>
          <g ref={linksRef} />
          <g ref={nodesRef} />
        </g>
      </svg>

      {hover && (
        <div
          className="pointer-events-none absolute z-20 max-w-xs -translate-y-full rounded-lg border border-edge bg-paper px-3 py-2 shadow-[0_8px_24px_-12px_rgba(28,24,18,0.5)]"
          style={{ left: hover.x + 12, top: hover.y - 8 }}
        >
          <p className="font-display text-[13px] text-ink">{hover.node.label}</p>
          <dl className="mt-1 space-y-0.5 font-mono text-[11px] text-ink-soft">
            <div className="flex gap-2">
              <dt className="w-16 shrink-0">type</dt>
              <dd className="text-ink">{hover.node.kind}</dd>
            </div>
            {hover.node.linkAddress && (
              <div className="flex gap-2">
                <dt className="w-16 shrink-0">link</dt>
                <dd className="text-ink">{fmtAddr(hover.node.linkAddress)}</dd>
              </div>
            )}
            {hover.node.loopBound && (
              <div className="flex gap-2">
                <dt className="w-16 shrink-0">loop</dt>
                <dd className="text-[#e6983f]">dead-end branch</dd>
              </div>
            )}
            {hover.node.redirect && (
              <div className="flex gap-2">
                <dt className="w-16 shrink-0">redirect</dt>
                <dd className="text-ink">{fmtAddr(hover.node.redirectAddress)}</dd>
              </div>
            )}
          </dl>
        </div>
      )}
    </div>
  )
}

function idOf(x: string | SimNode): string {
  return typeof x === 'string' ? x : x.id
}

// Shortest hop-distance from the Philosophy node, over undirected edges.
function bfsFromPhilosophy(
  nodes: GraphNodeLite[],
  links: GraphLinkLite[],
): Record<string, number> {
  const dist: Record<string, number> = {}
  const start = nodes.find((n) => n.kind === 'philosophy')?.id
  if (!start) return dist

  const adj: Record<string, string[]> = {}
  for (const l of links) {
    ;(adj[l.source] ??= []).push(l.target)
    ;(adj[l.target] ??= []).push(l.source)
  }

  const queue = [start]
  dist[start] = 0
  while (queue.length) {
    const cur = queue.shift()!
    for (const nb of adj[cur] ?? []) {
      if (dist[nb] == null) {
        dist[nb] = dist[cur] + 1
        queue.push(nb)
      }
    }
  }
  return dist
}

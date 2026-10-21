// One entry in the Wikipedia search response (data.query.search[]).
export type SearchResult = {
  title: string
  pageid: number
  size: number
  wordcount: number
  snippet: string
  timestamp: string
}

export type TitlesResponse = {
  query?: { search?: SearchResult[] }
}

// An article the user has chosen to trace a path for.
export type Article = {
  pageid: number
  title: string
}

// Crawl mode sent to the backend.
export type CrawlMode = 'relaxed' | 'strict'

// The `value` object carried by a `node` SSE event (mirrors the backend linkDTO).
export type WireLink = {
  title: string
  linkAddress?: string
  redirect?: boolean
  redirectAddress?: string
  loopBound?: boolean
}

// ---- live graph model, built incrementally from the SSE stream ----

export type NodeKind = 'seed' | 'link' | 'loop' | 'philosophy'

export type GraphNodeLite = {
  id: string
  label: string
  kind: NodeKind
  // node properties (from the event `value`, plus later redirect/loop events) — shown on hover
  linkAddress?: string
  redirect?: boolean
  redirectAddress?: string
  loopBound?: boolean
}

export type GraphLinkLite = {
  source: string
  target: string
}

// Per-seed-article progress, surfaced in the sidebar.
export type ChainInfo = {
  path: string[] // ordered node ids, starting with the seed (main path only)
  done: boolean // reached Philosophy
  looped: boolean
  loops?: string[] // dead-end nodes pruned to loop branches
  error?: string
}

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

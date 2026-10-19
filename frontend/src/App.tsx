import { useState } from 'react'
import type { Article } from './types'
import { SearchPanel } from './components/SearchPanel'
import { Sidebar } from './components/Sidebar'
import { GraphCanvas } from './components/GraphCanvas'

const MAX = 30

export default function App() {
  const [selected, setSelected] = useState<Article[]>([])
  const [crawled, setCrawled] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(false)

  function addArticle(a: Article) {
    setSelected((prev) => {
      if (prev.length >= MAX || prev.some((p) => p.pageid === a.pageid)) return prev
      return [...prev, a]
    })
  }

  function removeArticle(pageid: number) {
    setSelected((prev) => prev.filter((p) => p.pageid !== pageid))
  }

  function findPaths() {
    if (selected.length === 0) return
    // Stub: the crawl endpoint (SSE) doesn't exist yet. For now this transitions
    // to the post-crawl layout (search hides) and slides the sidebar in.
    setCrawled(true)
    setSidebarOpen(true)
  }

  function backToSearch() {
    setCrawled(false)
    setSidebarOpen(false)
  }

  return (
    <div className="h-full p-3 md:p-5">
      <div className="relative mx-auto flex h-full max-w-350 gap-3 md:gap-4">
        {sidebarOpen && (
          <>
            {/* mobile-only scrim behind the drawer */}
            <div
              onClick={() => setSidebarOpen(false)}
              className="fixed inset-0 z-30 bg-ink/30 md:hidden"
              aria-hidden="true"
            />
            <Sidebar
              selected={selected}
              onRemove={removeArticle}
              crawled={crawled}
              max={MAX}
              onBackToSearch={backToSearch}
              onClose={() => setSidebarOpen(false)}
            />
          </>
        )}

        <div className="flex min-w-0 flex-1 flex-col gap-3 md:gap-4">
          {!crawled && (
            <SearchPanel
              selected={selected}
              onAdd={addArticle}
              onRemove={removeArticle}
              onFindPaths={findPaths}
              max={MAX}
            />
          )}
          <GraphCanvas
            crawled={crawled}
            sidebarOpen={sidebarOpen}
            onToggleSidebar={() => setSidebarOpen((v) => !v)}
          />
        </div>
      </div>
    </div>
  )
}

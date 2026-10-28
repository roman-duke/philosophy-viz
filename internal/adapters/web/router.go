package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/roman-duke/philosophy-viz/internal/core"
)

//go:embed dist
var distFS embed.FS

type CrawlRequest struct {
	Mode     core.TraverseMode `json:"mode"`
	Articles []string          `json:"articles"`
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// For now, we have just one route
	mux.Handle("GET /titles", AdapterHandler(h.getTitles))
	mux.Handle("POST /crawl", AdapterHandler(h.getPaths))

	// handle the serving of static assets
	distSub, _ := fs.Sub(distFS, "dist")
	mux.Handle("/", http.FileServer(http.FS(distSub)))

	return withLogging(mux)
}

package web

import (
	"net/http"

	"github.com/roman-duke/philosophy-viz/internal/core"
)

type CrawlRequest struct {
	Mode     core.TraverseMode `json:"mode"`
	Articles []string          `json:"articles"`
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// For now, we have just one route
	mux.Handle("GET /titles", AdapterHandler(h.getTitles))
	mux.Handle("POST /crawl", AdapterHandler(h.getPaths))

	return withLogging(mux)
}

package web

import "net/http"

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// For now, we have just one route
	mux.Handle("GET /titles", AdapterHandler(h.getTitles))

	return withLogging(mux)
}

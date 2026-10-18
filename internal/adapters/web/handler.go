package web

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/roman-duke/philosophy-viz/internal/core"
)

type Handler struct {
	svc *core.Service
}

var (
	ErrMalFormedQueryParams = errors.New("keyword key in query is missing")
	ErrWikiQuery            = errors.New("could not get search result from wikipedia")
)

type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e HTTPError) Error() string {
	return e.Err.Error()
}

func NewHandler(s *core.Service) *Handler {
	return &Handler{svc: s}
}

// handler to get a list of Wiki article titles via external fetch call
// to wikipedia api that fuzzy searches
func (h *Handler) getTitles(w http.ResponseWriter, r *http.Request) error {
	// read the query and extract the keyword that the client intends to search
	key := r.URL.Query().Get("keyword")

	if len(key) == 0 {
		return ErrMalFormedQueryParams
	}

	// now, we call the service with this keyword
	result, err := h.svc.FuzzyFindTitles(key)

	if err != nil {
		return ErrWikiQuery
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	return nil
}

// func (h *Handler) getPaths(w http.ResponseWriter, r *http.Request) {
// }

// ======================================== ERROR HANDLER ========================================//
type AdapterHandler func(http.ResponseWriter, *http.Request) error

func (h AdapterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}

	code, msg := http.StatusInternalServerError, "internal server error"

	var he HTTPError
	switch {
	case errors.As(err, &he):
	case errors.Is(err, ErrMalFormedQueryParams):
		code, msg = http.StatusBadRequest, "query params invalid"
	case errors.Is(err, ErrWikiQuery):
		code, msg = http.StatusInternalServerError, "external fetch request failed"
	}

	log.Printf("handler error: %v", err)
	http.Error(w, msg, code) // sets status + write body, in the right order
}

// Relevant middlewares
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// we call the original handler at this point
		next.ServeHTTP(w, r)

		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

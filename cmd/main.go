package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/roman-duke/philosophy-viz/internal/adapters/store"
	"github.com/roman-duke/philosophy-viz/internal/adapters/web"
	"github.com/roman-duke/philosophy-viz/internal/core"
)

func main() {
	store, err := store.NewStorage("file:cache.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")

	if err != nil {
		log.Fatalf("\033[31m%v\033[0m", err)
		os.Exit(1)
	}

	svc := core.NewService(store)
	h := web.NewHandler(svc)

	fmt.Printf("\033[35mServer listening at http://localhost:%v\033[0m\n", 8080)

	err = http.ListenAndServe(":8080", h.Routes())

	if err != nil {
		log.Fatalf("\033[31m%v\033[0m", err)
		os.Exit(1)
	}
}

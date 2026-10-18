package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/roman-duke/philosophy-viz/internal/adapters/web"
	"github.com/roman-duke/philosophy-viz/internal/core"
)

func main() {
	svc := core.NewService(nil)
	h := web.NewHandler(svc)

	fmt.Printf("\033[35mServer listening at http://localhost:%v\033[0m\n", 8080)

	err := http.ListenAndServe(":8080", h.Routes())

	if err != nil {
		log.Fatalf("\033[32m%v\033[0m", err)
		os.Exit(1)
	}
}

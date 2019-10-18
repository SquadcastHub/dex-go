package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/squadcastHQ/dex-go/pkg/dex"
)

func main() {

	r := chi.NewRouter()
	d := dex.New("<Your API Key here>")
	r.Use(d.Middleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello")
	})

	http.ListenAndServe(":9091", r)
}

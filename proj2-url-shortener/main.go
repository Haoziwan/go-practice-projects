package main

import (
	"log"
	"net/http"
)

func main() {
	// Initialize handlers and templates
	InitHandlers()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			HomeHandler(w, r)
		} else {
			RedirectHandler(w, r)
		}
	})

	http.HandleFunc("/shorten", ShortenHandler)

	// Start server
	port := ":8080"
	log.Println("Server starting on http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

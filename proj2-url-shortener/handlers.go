package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

var (
	templates *template.Template
	store     *URLStore
)

// InitHandlers initializes handlers and templates
func InitHandlers() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Failed to parse templates:", err)
	}

	store = NewURLStore()
}

// HomeHandler shows the main form page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		err := templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Println("Template error:", err)
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ShortenHandler processes URL shortening requests
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(r.FormValue("url"))
	if longURL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	// Add http:// if no scheme is present
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "http://" + longURL
	}

	// Generate short key
	shortKey, err := store.GenerateShortKey()
	if err != nil {
		http.Error(w, "Failed to generate short URL", http.StatusInternalServerError)
		log.Println("Generate key error:", err)
		return
	}

	// Save mapping
	err = store.Save(shortKey, longURL)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		log.Println("Save error:", err)
		return
	}

	// Build the full short URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	shortURL := scheme + "://" + r.Host + "/" + shortKey

	// Render result page
	data := map[string]string{
		"ShortURL": shortURL,
		"LongURL":  longURL,
	}

	err = templates.ExecuteTemplate(w, "result.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Println("Template error:", err)
	}
}

// RedirectHandler redirects short URLs to long URLs
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortKey := strings.TrimPrefix(r.URL.Path, "/")

	if shortKey == "" {
		http.NotFound(w, r)
		return
	}

	longURL, err := store.Get(shortKey)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Use StatusFound (302) for temporary redirect
	// This allows the short URL service to track analytics if needed
	http.Redirect(w, r, longURL, http.StatusFound)
}

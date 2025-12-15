package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
func HomeHandler(c *gin.Context) {
	err := templates.ExecuteTemplate(c.Writer, "index.html", nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render template")
		log.Println("Template error:", err)
	}
}

// ShortenHandler processes URL shortening requests
func ShortenHandler(c *gin.Context) {
	longURL := strings.TrimSpace(c.PostForm("url"))
	if longURL == "" {
		c.String(http.StatusBadRequest, "URL cannot be empty")
		return
	}

	// Add http:// if no scheme is present
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "http://" + longURL
	}

	// Generate short key
	shortKey, err := store.GenerateShortKey()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate short URL")
		log.Println("Generate key error:", err)
		return
	}

	// Save mapping
	err = store.Save(shortKey, longURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save URL")
		log.Println("Save error:", err)
		return
	}

	// Build the full short URL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	shortURL := scheme + "://" + c.Request.Host + "/" + shortKey

	// Render result page
	data := map[string]string{
		"ShortURL": shortURL,
		"LongURL":  longURL,
	}

	err = templates.ExecuteTemplate(c.Writer, "result.html", data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render template")
		log.Println("Template error:", err)
	}
}

// RedirectHandler redirects short URLs to long URLs
func RedirectHandler(c *gin.Context) {
	shortKey := c.Param("shortKey")

	if shortKey == "" {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	longURL, err := store.Get(shortKey)
	if err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	// Use StatusFound (302) for temporary redirect
	// This allows the short URL service to track analytics if needed
	c.Redirect(http.StatusFound, longURL)
}

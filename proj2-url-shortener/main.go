package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize handlers and templates
	InitHandlers()

	// Create Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Route handlers
	r.GET("/", HomeHandler)
	r.POST("/shorten", ShortenHandler)
	r.GET("/:shortKey", RedirectHandler)

	// Start server
	port := ":8080"
	log.Println("Server starting on http://localhost" + port)
	log.Fatal(r.Run(port))
}

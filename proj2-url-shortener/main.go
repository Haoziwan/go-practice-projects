package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Initialize handlers and templates
	InitHandlers()

	// Test Redis connection
	if err := store.Ping(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Connected to Redis successfully")

	// Ensure Redis connection is closed on exit
	defer store.Close()

	// Create Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Route handlers
	r.GET("/", HomeHandler)
	r.POST("/shorten", ShortenHandler)
	r.GET("/:shortKey", RedirectHandler)

	// Get server port from environment variable or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	// Start server
	log.Println("Server starting on http://localhost" + port)
	log.Fatal(r.Run(port))
}

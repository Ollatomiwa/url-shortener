package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // SQLite driver
)

// step4: storage: map short codes to original url
var urlStore = make(map[string]string)

// step4: global random generator
var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))

// step4: helper:  generate a random short code (6 characters)
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	//step4: avoid duplicate
	for {
		b := make([]byte, 10) // length of short code
		for i := range b {
			b[i] = charset[randGen.Intn(len(charset))]
		}
		shortCode := string(b)
		if _, exists := urlStore[shortCode]; !exists {
			return shortCode // return only if unique
		}
	}
}

// step3: define a struct for the request
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// step6: initialize the database
func initDB() *sql.DB {
	// Use environment variable for database path or default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "file:urlshortener.db?cache=shared"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create the table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (short_code TEXT PRIMARY KEY, original_url TEXT NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	return db
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins or all for testing
		allowedOrigins := []string{
			"https://cplshort.vercel.app",
			"http://localhost:5173",
			"http://localhost:8080",
		}

		allowed := false
		for _, o := range allowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Simple CORS middleware alternative (uncomment if you want to allow all origins)
func SimpleCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	// Get port from environment variable (for Railway/Render deployment)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := initDB()   //step6: Initialize database
	defer db.Close() //step6: Close DB connection

	// step1: initialize Gin Router (with default middleware: Logger and Recovery)
	r := gin.Default()

	// Use CORS middleware - CHOOSE ONE OF THE FOLLOWING:

	// Option 1: Use the configurable CORS middleware (recommended for production)
	r.Use(CORSMiddleware())

	// Option 2: Uncomment below to allow all origins (for testing)
	// r.Use(SimpleCORSMiddleware())

	//step1: define a basic route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to the URL shortener API"})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "timestamp": time.Now()})
	})

	//step2: adding a new endpoint to shorten a url
	r.POST("/shorten", func(c *gin.Context) {
		//step3: updating /shorten handler
		var req ShortenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Println("Binding error:", err)
			c.JSON(400, gin.H{"error": "Invalid Request Body: " + err.Error()})
			return
		}
		fmt.Println("Received URL:", req.URL)

		// Validate URL format
		if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
			c.JSON(400, gin.H{"error": "URL must start with http:// or https://"})
			return
		}

		//step4: modify to store the url and return the short code
		shortCode := generateShortCode()

		//step6: Replace the in-memory map with db operations
		_, err := db.Exec("INSERT INTO urls(short_code, original_url) VALUES (?, ?)", shortCode, req.URL)
		if err != nil {
			fmt.Println("Database error:", err)
			c.JSON(500, gin.H{"error": "Failed to save URL: " + err.Error()})
			return
		}

		// Get base URL for the short link
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = port + ".cpl"
		}

		shortURL := fmt.Sprintf("%s/%s", baseURL, shortCode)

		c.JSON(200, gin.H{
			"original_url": req.URL,
			"short_url":    shortURL,
			"short_code":   shortCode,
		})
	})

	//step4: adding a debug endpoint to view all mappings
	r.GET("/debug", func(c *gin.Context) {
		rows, err := db.Query("SELECT short_code, original_url FROM urls ORDER BY created_at DESC LIMIT 100")
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to query URLs"})
			return
		}
		defer rows.Close()

		urls := make(map[string]string)
		for rows.Next() {
			var shortCode, originalURL string
			if err := rows.Scan(&shortCode, &originalURL); err != nil {
				continue
			}
			urls[shortCode] = originalURL
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)

		c.JSON(200, gin.H{
			"count":    count,
			"mappings": urls,
		})
	})

	//step4: adding redirection endpoint
	r.GET("/:shortCode", func(c *gin.Context) {
		shortCode := c.Param("shortCode")

		// Remove .cpl extension if present
		shortCode = strings.TrimSuffix(shortCode, ".cpl")

		var originalURL string
		err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = ?", shortCode).Scan(&originalURL)
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Short URL not found"})
			return
		} else if err != nil {
			c.JSON(500, gin.H{"error": "Database Error: " + err.Error()})
			return
		}

		// redirect to the original url
		c.Redirect(302, originalURL)
	})

	//step1: Running the server on Port
	log.Printf("Server starting on port %s", port)
	log.Printf("Allowed origins: %s", os.Getenv("ALLOWED_ORIGINS"))

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

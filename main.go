package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
	"github.com/gin-gonic/gin"
    _ "modernc.org/sqlite" // SQLite driver
)

//step4: storage: map short codes to original url
var urlStore = make(map[string]string)

//step4: global random generator
var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))



//step4: helper:  generate a random short code (6 characters)
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	//step4: avoid duplicate
	for {
		b := make([]byte, 6)
		for i := range b {
			b[i] = charset[randGen.Intn(len(charset))]
		}
		shortCode := string(b)
		if _, exists := urlStore[shortCode]; !exists {

			return shortCode // return only if unique
		}

	}
}

//step3: define a struct for the request
type ShortenRequest struct{
	URL string `json:"url" binding:"required"` 
}

//step6: initialize the database
func initDB() *sql.DB {
	db, err := sql.Open("sqlite", "file:urlshortener.db? cache=shared")
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
func main () {
	db := initDB() //step6: Initialize database
	
	defer db.Close() //step6: Close DB connection
	

	// step1: initialize Gin Router (with default middleware: Logger and Recovery)
	r := gin.Default()
	
	//step1: define a basic route
	r.GET("/", func (c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to the URL shortener",})
	})

	//step2: adding a new endpoint to shorten a url 
	r.POST("/shorten", func (c *gin.Context)  {
		//step3: updating /shorten handler
		var req ShortenRequest
		if err := c.ShouldBindJSON(&req);
		err != nil {
			fmt.Println("Binding error:", err)
			 c.JSON(400, gin.H{"error": "Invalid Request Body"})
		return
		} 
		fmt.Println("Received URL:", req.URL) // Debug 2

		//step4: modify to store the url and return the short code
		shortCode := generateShortCode() 
		//step6: Replace the in-memory map with db operations
		_, err := db.Exec("INSERT INTO urls(short_code, original_url) VALUES (?, ?)", shortCode, req.URL,)
		if err != nil {
			c.JSON(500, gin.H{"eror": "failed to save URL"})
			return
		}

		c.JSON(200, gin.H{"original_url": req.URL, "short_url": "http://localhost:8080/" + shortCode, }) //full short url
	})
	
	//step4: adding a debug endpoint to view all mappings
	r.GET("/debug", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"count":    len(urlStore),
			"mappings": urlStore,
		})
	})

	//step4: adding redirection endpoint
	r.GET("/:shortCode", func(c *gin.Context) {
		shortCode := c.Param("shortCode") //step6: modify the GET /:shortCode handler to query the db
		var originalURL string
		err := db.QueryRow("SELECT original_url FROM urls WHERE short_code = ?", shortCode,).Scan(&originalURL)
		if err == sql.ErrNoRows {c.JSON(404, gin.H{"error":"Short URL not found"})
		return
	} else if err !=nil {c.JSON(500, gin.H{"error": "Database Error"})
		return
	}
		// redirect to the original url
		c.Redirect(302, originalURL)
	})
	//step1: Running the server on Port 8080
	r.Run(":8080")
}
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
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

func main () {


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
		urlStore[shortCode] = req.URL // store the mapping 

		fmt.Println("Stored mapping:", shortCode, "->", urlStore[shortCode]) // Debug 3

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
		shortCode := c.Param("shortCode")
		originalURL, exists := urlStore[shortCode]
		if !exists {
			c.JSON(404, gin.H{"error": "Short URL not found"})
			return
		}
		// redirect to the original url
		c.Redirect(302, originalURL)
	})
	//step1: Running the server on Port 8080
	r.Run(":8080")
}
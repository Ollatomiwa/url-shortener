package main

import (
	"github.com/gin-gonic/gin"
)

//step3: define a struct for the request
type ShortenRequest struct{
	URL string `json:"url" binding:"required` 
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
		if err := c.ShouldBindBodyWithJSON(&req);
		err != nil { c.JSON(400, gin.H{"error": "Invalid Request Body"})
		return
		} 

		c.JSON(200, gin.H{"original_url": req.URL, "shorten_url": "Todo: generate short url",})
	})
	
	//step1: Running the server on Port 8080
	r.Run(":8080")
}
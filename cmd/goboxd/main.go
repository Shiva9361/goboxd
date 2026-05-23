package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hello, Goboxd!")
	router := gin.Default()
	router.GET("healthz", getHealth)
	router.Run(":8080")
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thesouldev/goboxd/internal"
)

func main() {
	log.Println("Starting goboxd...")

	code_runner := new(internal.CodeRunner)

	code_runner.Initialize()
	
	router := gin.Default()
	
	// Registering endpoints
	router.GET("healthz", getHealth)
	router.POST("run", postRun)

	router.Run(":8080")
}

func getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func postRun(c *gin.Context) {
	
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
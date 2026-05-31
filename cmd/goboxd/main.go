package main

import (
	"log"
	"os"

	"github.com/thesouldev/goboxd/internal"
)

func main() {
	server := internal.NewServer()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := server.Router.Run(":" + port); err != nil {
		log.Fatalln("Could not start server due to: ", err)
	}

}

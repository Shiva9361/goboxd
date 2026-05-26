package main

import (
	"log"

	"github.com/thesouldev/goboxd/internal"
)

func main() {
	server := internal.NewServer()
	if err := server.Router.Run(); err != nil {
		log.Fatalln("Could not start server due to: ", err)
	}

}

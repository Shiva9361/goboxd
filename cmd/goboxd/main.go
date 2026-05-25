package main

import (
	"log"

	"github.com/thesouldev/goboxd/internal"
)

func main() {
	app := internal.NewServer()
	if err := app.Router.Run(); err != nil {
		log.Fatalln("Could not start app due to: ", err)
	}

}

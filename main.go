package main

import (
	"log"

	"github.com/mateconpizza/gmweb/internal/application"
)

var version = "0.1.1"

func main() {
	app := application.New(version)
	if err := app.Parse(); err != nil {
		log.Fatal(err)
	}

	log.Fatal(run(app))
}

package main

import (
	"log"

	"github.com/mateconpizza/gmweb/internal/application"
)

func main() {
	app := application.New().Parse()
	log.Fatal(run(app))
}

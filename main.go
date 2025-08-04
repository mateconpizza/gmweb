package main

import (
	"embed"
	"log"

	"github.com/mateconpizza/gmweb/internal/application"
)

var (
	//go:embed ui/templates/**/*.gohtml
	templates embed.FS

	//go:embed ui/static/**/*
	staticFiles embed.FS
)

func main() {
	app := application.New().Parse()
	app.Server.TemplatesFiles = &templates
	app.Server.StaticFiles = &staticFiles
	app.Server.CertFile = "./tls/localhost.pem"
	app.Server.KeyFile = "./tls/localhost-key.pem"

	log.Fatal(run(app))
}

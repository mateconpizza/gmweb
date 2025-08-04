// Package application contains the configuration for the application.
package application

import (
	"fmt"
	"os"
)

type App struct {
	Cfg    *Config
	Flags  *Flags
	Server *Server
}

func New() *App {
	return &App{
		Cfg: &Config{
			Name:    appName,
			MainDB:  mainDB,
			DataDir: "gomarks",
			Info: &information{
				URL:   "https://github.com/mateconpizza/gmweb#readme",
				Title: "A simple web bookmark manager",
				Tags:  "golang,awesome,bookmarks,cli,manager",
				Desc:  "Simple yet powerful bookmark manager for your browser",
			},
		},
		Flags: &Flags{
			Addr: ":8080",
		},
		Server: &Server{
			QRImgSize:    512,
			ItemsPerPage: 32,
		},
	}
}

func (a *App) Usage() {
	fmt.Fprintf(os.Stderr, `%s
%s

Usage:
  %s [options]

Options:
  -p, --path <path>	Path to store data (default: %s)
  -a, --addr <addr>	Address to listen on (default: %s)
  -v, --verbose		Increase verbosity (-v, -vv, -vvv)
  -V, --version		Show version
  -h, --help		Show this help
`, a.Cfg.String(), a.Cfg.Info.Title, a.Cfg.Name, a.Flags.Path, a.Flags.Addr)
}

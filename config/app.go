// Package config contains the configuration for the application.
package config

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	gap "github.com/muesli/go-app-paths"
	flag "github.com/spf13/pflag"
)

var version string = "0.1.0"

const (
	appName    string = "gmweb"
	mainDBName string = "main.db" // Default name of the main database
)

type (
	Config struct {
		Name    string       `json:"name"` // Name of the application
		DataDir string       `json:"data"` // Data directory
		MainDB  string       `json:"db"`   // Database name
		Info    *information `json:"info"` // Application information
		Server  *Server      `json:"-"`    // Server configuration
		Flags   *Flags       `json:"-"`    // Flags app CLI flags
	}

	Server struct {
		QRImgSize      int       // QR image size
		ItemsPerPage   int       // ItemsPerPage
		TemplatesFiles *embed.FS // Templates
		StaticFiles    *embed.FS // Static files
		CertFile       string    // Certificate file
		KeyFile        string    // Key file
	}

	Flags struct {
		Path    string
		Addr    string
		Verbose int
		Version bool
	}

	information struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Tags  string `json:"tags"`
		Desc  string `json:"desc"`
	}
)

func (c *Config) String() string {
	return fmt.Sprintf("%s v%s %s/%s\n", c.Name, version, runtime.GOOS, runtime.GOARCH)
}

// dataPath returns the data path for the application.
func dataPath(c *Config) (string, error) {
	scope := gap.NewScope(gap.User, c.DataDir)

	dataDir, err := scope.DataPath("")
	if err != nil {
		return "", fmt.Errorf("getting data path: %w", err)
	}

	return dataDir, nil
}

func Usage(c *Config) {
	fmt.Printf(`%s
%s

Usage:
  %s [options]

Options:
  -p, --path <path>	Path to store data (default: %s)
  -a, --addr <addr>	Address to listen on (default: %s)
  -v, --verbose		Increase verbosity (-v, -vv, -vvv)
  -V, --version		Show version
  -h, --help		Show this help
`, c.String(), c.Info.Title, c.Name, c.Flags.Path, c.Flags.Addr)
}

func New() *Config {
	return &Config{
		Name:    appName,
		MainDB:  mainDBName,
		DataDir: "gomarks",
		Info: &information{
			URL:   "https://github.com/mateconpizza/gmweb#readme",
			Title: "A simple web bookmark manager",
			Tags:  "golang,awesome,bookmarks,cli,manager",
			Desc:  "Simple yet powerful bookmark manager for your webbrowser",
		},
		Server: &Server{
			QRImgSize:    512,
			ItemsPerPage: 32,
		},
		Flags: &Flags{
			Addr: ":8080",
		},
	}
}

func Parse(c *Config) {
	s, err := dataPath(c)
	if err != nil {
		panic(err)
	}

	flag.StringVarP(&c.Flags.Path, "path", "p", s, "")
	flag.StringVarP(&c.Flags.Addr, "addr", "a", c.Flags.Addr, "")
	flag.CountVarP(&c.Flags.Verbose, "verbose", "v", "Increase verbosity (-v, -vv, -vvv)")
	flag.BoolVarP(&c.Flags.Version, "version", "V", false, "")
	flag.Usage = func() {
		Usage(c)
	}
	flag.Parse()

	setVerbosity(c.Flags.Verbose)
}

func setVerbosity(verbose int) {
	levels := []slog.Level{
		slog.LevelError,
		slog.LevelWarn,
		slog.LevelInfo,
		slog.LevelDebug,
	}
	level := levels[min(verbose, len(levels)-1)]

	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "source" {
					if source, ok := a.Value.Any().(*slog.Source); ok {
						dir, file := filepath.Split(source.File)
						source.File = filepath.Join(filepath.Base(filepath.Clean(dir)), file)

						return slog.Attr{Key: "source", Value: slog.AnyValue(source)}
					}
				}

				return a
			},
		}),
	)
	slog.SetDefault(logger)

	slog.Debug("logging", "level", level)
}

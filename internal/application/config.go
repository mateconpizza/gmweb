package application

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mateconpizza/gm/pkg/files"
	gap "github.com/muesli/go-app-paths"
	flag "github.com/spf13/pflag"
)

var version string = "0.1.0"

const (
	appName string = "gmweb"
	mainDB  string = "main" // Default name of the main database
)

type (
	// Config holds the overall application configuration.
	Config struct {
		Name     string       `json:"name"`  // Name of the application
		DataDir  string       `json:"data"`  // Data directory
		CacheDir string       `json:"cache"` // Cache data directory
		MainDB   string       `json:"db"`    // Database name
		Info     *information `json:"info"`  // Application information
	}

	// Server holds configuration specific to the web server.
	Server struct {
		QRImgSize    int    // QR image size
		ItemsPerPage int    // ItemsPerPage
		CertFile     string // Certificate file path for HTTPS
		KeyFile      string // Key file path for HTTPS
	}

	// Flags holds command-line interface flags.
	Flags struct {
		Path    string // Path to store data
		Addr    string // Address to listen on
		DevMode bool   // Development mode
		Verbose int    // Verbosity
		Version bool   // Version
		Help    bool   // Help
	}

	// information holds general application metadata.
	information struct {
		Author    string `json:"author"`
		AuthorURL string `json:"author_url"`
		URL       string `json:"url"`
		Title     string `json:"title"`
		Tags      string `json:"tags"`
		Desc      string `json:"desc"`
	}
)

func (a *Config) String() string {
	return fmt.Sprintf("%s v%s %s/%s\n", a.Name, version, runtime.GOOS, runtime.GOARCH)
}

// dataPath load data and cache directories.
func (a *App) loadPaths() error {
	scope := gap.NewScope(gap.User, a.Cfg.DataDir)

	// databases, config files.
	dataDir, err := scope.DataPath("")
	if err != nil {
		return fmt.Errorf("getting data path: %w", err)
	}

	// web files, favicons, etc.
	cacheDir, err := scope.CacheDir()
	if err != nil {
		return fmt.Errorf("getting cache path: %w", err)
	}

	a.Cfg.DataDir = dataDir
	a.Cfg.CacheDir = cacheDir

	return nil
}

func (a *App) Parse() *App {
	if err := a.loadPaths(); err != nil {
		panic(err)
	}

	flag.StringVarP(&a.Flags.Path, "path", "p", a.Cfg.DataDir, "")
	flag.StringVarP(&a.Flags.Addr, "addr", "a", a.Flags.Addr, "")
	flag.BoolVarP(&a.Flags.DevMode, "dev", "d", false, "")
	flag.CountVarP(&a.Flags.Verbose, "verbose", "v", "Increase verbosity (-v, -vv, -vvv)")
	flag.BoolVarP(&a.Flags.Version, "version", "V", false, "")
	flag.BoolVarP(&a.Flags.Help, "help", "h", false, "")
	flag.Parse()

	if a.Flags.Version {
		fmt.Print(a.Cfg.String())
		os.Exit(0)
	}

	if a.Flags.Help {
		a.Usage()
		os.Exit(1)
	}

	a.Cfg.DataDir = a.Flags.Path
	if err := files.MkdirAll(a.Cfg.CacheDir, a.Cfg.DataDir); err != nil {
		panic(err)
	}

	return a
}

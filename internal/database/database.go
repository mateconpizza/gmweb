package database

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mateconpizza/gmweb/internal/models"
)

var ErrDBNotAllowed = errors.New("database not allowed")

var (
	Valid       = make(map[string]string)
	connections = make(map[string]models.Repo)
	mu          sync.RWMutex
)

// Get returns a reusable connection to the requested base.
func Get(dbKey string) (models.Repo, error) {
	mu.RLock()
	repo, exists := connections[dbKey]
	mu.RUnlock()

	if exists {
		slog.Debug("database connection already open", "database", dbKey)
		return repo, nil
	}

	path, ok := Path(dbKey)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrDBNotAllowed, dbKey)
	}

	dsn := path
	newDB, err := models.New(dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database %s: %w", dbKey, err)
	}

	mu.Lock()
	connections[dbKey] = newDB
	mu.Unlock()

	return newDB, nil
}

// CloseAll closes all connections.
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()

	slog.Info("database: closing connections")
	if len(connections) == 0 {
		slog.Info("database: no connections found")
		return
	}

	for key, db := range connections {
		db.Close()
		delete(connections, key)
	}
}

// IsValid verify if an identifier is allowed.
func IsValid(dbKey string) bool {
	_, ok := Valid[dbKey]
	return ok
}

// Path returns the database file route according to the identifier.
func Path(dbKey string) (string, bool) {
	path, ok := Valid[dbKey]
	return path, ok
}

func Register(dbKey, dbPath string) {
	Valid[dbKey] = dbPath
}

func Forget(dbKey string) {
	delete(connections, dbKey)
	delete(Valid, dbKey)
}

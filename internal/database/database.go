package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mateconpizza/gmweb/internal/models"
)

var ErrDBNotAllowed = errors.New("database not allowed")

var (
	Valid       = make(map[string]string)
	connections = make(map[string]*models.BookmarkModel)
	mu          sync.RWMutex
)

// Get returns a reusable connection to the requested base.
func Get(dbKey string) (*models.BookmarkModel, error) {
	mu.RLock()
	repo, exists := connections[dbKey]
	mu.RUnlock()

	if exists {
		return repo, nil
	}

	path, ok := Path(dbKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDBNotAllowed, dbKey)
	}

	dsn := fmt.Sprintf("file:%s?_cache=shared&_journal_mode=WAL", path)
	newDB, err := models.New(dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir base de datos %s: %w", dbKey, err)
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

func Add(dbKey, dbPath string) {
	Valid[dbKey] = dbPath
}

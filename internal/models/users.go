package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid user credentials")
	ErrUserDuplicated     = errors.New("username already exists")
)

var users = map[string]User{
	"user": {
		ID:             1,
		Name:           "user",
		HashedPassword: []byte("pass"),
		CreatedAt:      time.Now(),
	},
	"other": {
		ID:             2,
		Name:           "other",
		HashedPassword: []byte("pass"),
		CreatedAt:      time.Now(),
	},
}

type User struct {
	ID             int
	Name           string
	HashedPassword []byte
	CreatedAt      time.Time
}

type UserModel struct {
	store *sql.DB
}

func (m *UserModel) Insert(name, password string) error {
	return nil
}

func (m *UserModel) Authenticate(name, password string) (int, error) {
	u, ok := users[name]
	if !ok {
		return 0, ErrInvalidCredentials
	}

	return u.ID, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}

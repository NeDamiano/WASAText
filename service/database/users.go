package database

import (
	"database/sql"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

// GetUserByName returns the user ID by name
func (db *appdbimpl) GetUserByName(name string) (int64, error) {
	var id int64
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", name).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return id, nil
}

// CreateUser creates a new user with the given name and returns its ID
func (db *appdbimpl) CreateUser(name string) (int64, error) {
	res, err := db.c.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

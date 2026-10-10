package database

import (
	"database/sql"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

// GetUserById returns the user with the given id
func (db *appdbimpl) GetUserById(id int64) (User, error) {
	var u User
	err := db.c.QueryRow("SELECT id, name, COALESCE(photo, '') FROM users WHERE id = ?", id).Scan(&u.Id, &u.Name, &u.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, ErrUserNotFound
		}
		return u, err
	}
	return u, nil
}

// GetUserByName returns the id of the user with the given name
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

// CreateUser creates a new user with the given name and returns its id
func (db *appdbimpl) CreateUser(name string) (int64, error) {
	res, err := db.c.Exec("INSERT INTO users (id, name, photo) VALUES (NULL, ?, NULL)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UserExists checks if the given id exists in the users table
func (db *appdbimpl) UserExists(id int64) (bool, error) {
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// SetUserName changes the username of the user
func (db *appdbimpl) SetUserName(id int64, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
	return err
}

// User in the database (no photo = "")
type User struct {
	Id    int64
	Name  string
	Photo string
}

// SearchUsers returns the users whose name contains text (no text = all users)
func (db *appdbimpl) SearchUsers(text string) ([]User, error) {
	rows, err := db.c.Query("SELECT id, name, COALESCE(photo, '') FROM users WHERE name LIKE ? ORDER BY name LIMIT 1000", "%"+text+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err = rows.Scan(&u.Id, &u.Name, &u.Photo)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return users, nil
}

// SetUserPhoto saves the URL of the user's photo
func (db *appdbimpl) SetUserPhoto(id int64, photo string) error {
	_, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", photo, id)
	return err
}

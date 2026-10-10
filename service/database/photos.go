package database

import (
	"database/sql"
	"errors"
)

var ErrPhotoNotFound = errors.New("photo not found")

// CreatePhoto saves the image bytes in the photos table and returns the new photo id
func (db *appdbimpl) CreatePhoto(data []byte) (int64, error) {
	res, err := db.c.Exec("INSERT INTO photos (id, data) VALUES (NULL, ?)", data)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetPhoto returns the image bytes of the photo with the given id
func (db *appdbimpl) GetPhoto(id int64) ([]byte, error) {
	var data []byte
	err := db.c.QueryRow("SELECT data FROM photos WHERE id = ?", id).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, err
	}
	return data, nil
}

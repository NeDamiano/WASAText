/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error
	GetUserByName(name string) (int64, error)
	CreateUser(name string) (int64, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			photo TEXT
		);

		CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY,	
			name TEXT,
			is_group INTEGER NOT NULL DEFAULT 0,
			photo TEXT,
			CHECK (is_group = 0 OR is_group = 1)
		);

		CREATE TABLE IF NOT EXISTS members (
			user_id INTEGER NOT NULL,
			conversation_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, conversation_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		);
		
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			text TEXT,
			photo TEXT,
			forwarded INTEGER NOT NULL DEFAULT 0,
			CHECK (text IS NOT NULL OR photo IS NOT NULL),
			CHECK (forwarded = 0 OR forwarded = 1),
			FOREIGN KEY (sender_id) REFERENCES users(id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		);

		CREATE TABLE IF NOT EXISTS message_status (
			user_id INTEGER NOT NULL,
			message_id INTEGER NOT NULL,
			status INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, message_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (message_id) REFERENCES messages(id),
			CHECK (status = 0 OR status = 1 OR status = 2)
		);

		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY,
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			emoji TEXT NOT NULL,
			FOREIGN KEY (message_id) REFERENCES messages(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

	`)
	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

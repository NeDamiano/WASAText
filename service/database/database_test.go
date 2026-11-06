package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// newTestDB returns a fresh, empty database that lives only in memory.
// Every test gets its own database, so tests don't affect each other.
func newTestDB(t *testing.T) AppDatabase {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		t.Fatalf("opening in-memory database: %v", err)
	}
	// an in-memory database exists only inside one connection
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = conn.Close() })

	db, err := New(conn)
	if err != nil {
		t.Fatalf("creating tables: %v", err)
	}
	return db
}

// mustCreateUser creates a user and stops the test if it fails.
func mustCreateUser(t *testing.T, db AppDatabase, name string) int64 {
	t.Helper()
	id, err := db.CreateUser(name)
	if err != nil {
		t.Fatalf("creating user %q: %v", name, err)
	}
	return id
}

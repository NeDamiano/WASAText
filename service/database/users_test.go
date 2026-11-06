package database

import (
	"errors"
	"testing"
)

func TestCreateAndGetUser(t *testing.T) {
	db := newTestDB(t)

	id := mustCreateUser(t, db, "alice")

	gotId, err := db.GetUserByName("alice")
	if err != nil {
		t.Fatalf("GetUserByName: %v", err)
	}
	if gotId != id {
		t.Errorf("GetUserByName returned id %d, want %d", gotId, id)
	}

	user, err := db.GetUserById(id)
	if err != nil {
		t.Fatalf("GetUserById: %v", err)
	}
	if user.Name != "alice" {
		t.Errorf("GetUserById returned name %q, want %q", user.Name, "alice")
	}
}

func TestUsernameMustBeUnique(t *testing.T) {
	db := newTestDB(t)

	mustCreateUser(t, db, "alice")

	_, err := db.CreateUser("alice")
	if err == nil {
		t.Error("creating a second user called alice should fail")
	}
}

func TestGetUnknownUser(t *testing.T) {
	db := newTestDB(t)

	_, err := db.GetUserByName("nobody")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("got error %v, want ErrUserNotFound", err)
	}
}

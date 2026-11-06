package database

import (
	"errors"
	"testing"
)

func TestDirectConversation(t *testing.T) {
	db := newTestDB(t)
	alice := mustCreateUser(t, db, "alice")
	bob := mustCreateUser(t, db, "bob")
	carol := mustCreateUser(t, db, "carol")

	convId, err := db.CreateDirectConversation(alice, bob)
	if err != nil {
		t.Fatalf("CreateDirectConversation: %v", err)
	}

	// the conversation is found in both directions
	got, err := db.GetDirectConversation(bob, alice)
	if err != nil {
		t.Fatalf("GetDirectConversation: %v", err)
	}
	if got != convId {
		t.Errorf("GetDirectConversation returned %d, want %d", got, convId)
	}

	// in a private chat, the name shown is the other user's name
	conv, err := db.GetConversation(convId, alice)
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if conv.Name != "bob" {
		t.Errorf("alice sees the chat named %q, want %q", conv.Name, "bob")
	}

	// a user who is not a member cannot see the conversation
	_, err = db.GetConversation(convId, carol)
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("carol got error %v, want ErrConversationNotFound", err)
	}
}

func TestConversationsSortedByLastMessage(t *testing.T) {
	db := newTestDB(t)
	alice := mustCreateUser(t, db, "alice")
	bob := mustCreateUser(t, db, "bob")
	carol := mustCreateUser(t, db, "carol")

	withBob, err := db.CreateDirectConversation(alice, bob)
	if err != nil {
		t.Fatalf("CreateDirectConversation: %v", err)
	}
	withCarol, err := db.CreateDirectConversation(alice, carol)
	if err != nil {
		t.Fatalf("CreateDirectConversation: %v", err)
	}

	if _, err = db.CreateMessage(withBob, bob, "first", "", 0, false); err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if _, err = db.CreateMessage(withCarol, carol, "second", "", 0, false); err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}

	convs, err := db.GetMyConversations(alice)
	if err != nil {
		t.Fatalf("GetMyConversations: %v", err)
	}
	if len(convs) != 2 {
		t.Fatalf("alice has %d conversations, want 2", len(convs))
	}
	// the chat with the newest message comes first
	if convs[0].Id != withCarol || convs[0].LastText != "second" {
		t.Errorf("first conversation is %d (%q), want %d (%q)", convs[0].Id, convs[0].LastText, withCarol, "second")
	}
}

package database

import (
	"testing"
)

// statusOf returns the status the sender sees for a message
// (0 sent, 1 received, 2 read).
func statusOf(t *testing.T, db AppDatabase, convId int64, msgId int64) int {
	t.Helper()
	msgs, err := db.GetMessages(convId)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	for _, m := range msgs {
		if m.Id == msgId {
			return m.Status
		}
	}
	t.Fatalf("message %d not found", msgId)
	return -1
}

// In a group the checkmarks follow the slowest member:
// one checkmark only when everyone received the message, two only when everyone read it.
func TestReadReceiptsInGroup(t *testing.T) {
	db := newTestDB(t)
	alice := mustCreateUser(t, db, "alice")
	bob := mustCreateUser(t, db, "bob")
	carol := mustCreateUser(t, db, "carol")

	group, err := db.CreateGroup("friends", []int64{alice, bob, carol})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	msg, err := db.CreateMessage(group, alice, "hi", "", 0, false)
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}

	steps := []struct {
		description string
		action      func() error
		wantStatus  int
	}{
		{"just sent", func() error { return nil }, 0},
		{"only bob received it", func() error { return db.MarkAsReceived(bob) }, 0},
		{"carol received it too", func() error { return db.MarkAsReceived(carol) }, 1},
		{"only bob read it", func() error { return db.MarkAsRead(group, bob) }, 1},
		{"carol read it too", func() error { return db.MarkAsRead(group, carol) }, 2},
	}

	for _, step := range steps {
		if err := step.action(); err != nil {
			t.Fatalf("%s: %v", step.description, err)
		}
		if got := statusOf(t, db, group, msg.Id); got != step.wantStatus {
			t.Errorf("%s: status %d, want %d", step.description, got, step.wantStatus)
		}
	}
}

func TestDeleteMessageRemovesComments(t *testing.T) {
	db := newTestDB(t)
	alice := mustCreateUser(t, db, "alice")
	bob := mustCreateUser(t, db, "bob")

	convId, err := db.CreateDirectConversation(alice, bob)
	if err != nil {
		t.Fatalf("CreateDirectConversation: %v", err)
	}
	msg, err := db.CreateMessage(convId, alice, "hello", "", 0, false)
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if _, err = db.CreateComment(msg.Id, bob, "👍"); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	if err = db.DeleteMessage(msg.Id); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}

	exists, err := db.MessageExists(msg.Id, convId)
	if err != nil {
		t.Fatalf("MessageExists: %v", err)
	}
	if exists {
		t.Error("the message still exists after DeleteMessage")
	}

	// comments are deleted together with the message (ON DELETE CASCADE)
	comments, err := db.GetComments(convId)
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("%d comments left after deleting the message, want 0", len(comments))
	}
}

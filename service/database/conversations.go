package database

import (
	"database/sql"
	"errors"
)

var ErrConversationNotFound = errors.New("conversation not found")

// GetDirectConversation returns the id of a conversation between 2 given users
func (db *appdbimpl) GetDirectConversation(user1, user2 int64) (int64, error) {
	var id int64
	err := db.c.QueryRow(`
		SELECT c.id FROM conversations c
		JOIN members m1 ON m1.conversation_id = c.id AND m1.user_id = ?
		JOIN members m2 ON m2.conversation_id = c.id AND m2.user_id = ?
		WHERE c.is_group = 0`, user1, user2).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrConversationNotFound
		}
		return 0, err
	}
	return id, nil
}

// CreateDirectConversation creates a direct conversation with user1 and user2
func (db *appdbimpl) CreateDirectConversation(user1 int64, user2 int64) (int64, error) {

	res, err := db.c.Exec("INSERT INTO conversations (id, is_group, name, photo) VALUES (NULL, 0, NULL, NULL)")
	if err != nil {
		return 0, err
	}

	convId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = db.c.Exec("INSERT INTO members (user_id, conversation_id) VALUES (?, ?), (?, ?)", user1, convId, user2, convId)
	if err != nil {
		return 0, err
	}

	return convId, nil
}

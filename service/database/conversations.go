package database

import (
	"database/sql"
	"errors"
)

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

// GetMyConversations returns the list of conversations of the logged in user
// (left joins so group chats aren't excluded while looking for the other user's name for private chats)
func (db *appdbimpl) GetMyConversations(userId int64) ([]ConversationSummary, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.is_group, COALESCE(c.name, u.name), COALESCE(c.photo, u.photo, ''),
			lastMes.timestamp, COALESCE(lastMes.text, ''), lastMes.photo IS NOT NULL
		FROM members m
		JOIN conversations c ON c.id = m.conversation_id
		LEFT JOIN members m2 ON m2.conversation_id = c.id AND m2.user_id != m.user_id AND c.is_group = 0
		LEFT JOIN users u ON u.id = m2.user_id
		LEFT JOIN messages lastMes ON lastMes.id = (SELECT id FROM messages WHERE conversation_id = c.id ORDER BY timestamp DESC LIMIT 1)
		WHERE m.user_id = ?
		ORDER BY lastMes.timestamp DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	convs := []ConversationSummary{}
	for rows.Next() {
		var conv ConversationSummary
		var lastTimestamp sql.NullTime
		err = rows.Scan(&conv.Id, &conv.IsGroup, &conv.Name, &conv.Photo,
			&lastTimestamp, &conv.LastText, &conv.LastIsPhoto)
		if err != nil {
			return nil, err
		}
		conv.HasLastMessage = lastTimestamp.Valid
		conv.LastTimestamp = lastTimestamp.Time
		convs = append(convs, conv)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return convs, nil
}

// GetConversation returns the logged-in user(userId)'s conversation(convId) summary
// (also for checking if a user is member of the conversation)
func (db *appdbimpl) GetConversation(convId int64, userId int64) (ConversationSummary, error) {
	var conv ConversationSummary
	err := db.c.QueryRow(`
		SELECT c.id, c.is_group, COALESCE(c.name, u.name), COALESCE(c.photo, u.photo, '')
		FROM members m
		JOIN conversations c ON c.id = m.conversation_id
		LEFT JOIN members m2 ON m2.conversation_id = c.id AND m2.user_id != m.user_id AND c.is_group = 0
		LEFT JOIN users u ON u.id = m2.user_id
		WHERE m.user_id = ? AND c.id = ?`, userId, convId).Scan(&conv.Id, &conv.IsGroup, &conv.Name, &conv.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return conv, ErrConversationNotFound
		}
		return conv, err
	}
	return conv, nil
}

// GetMembers returns the members of a conversation
func (db *appdbimpl) GetMembers(convId int64) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.name, COALESCE(u.photo, '')
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.conversation_id = ?
		ORDER BY u.name`, convId)
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

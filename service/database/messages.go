package database

import (
	"database/sql"
	"errors"
	"time"
)

// CreateMessage saves a new message and a "sent" status (0) for the other members
func (db *appdbimpl) CreateMessage(convId int64, senderId int64, text string, photo string, replyTo int64, forwarded bool) (Message, error) {
	timestamp := time.Now().UTC()

	// empty text, photo or replyTo are saved as NULL
	res, err := db.c.Exec(`
		INSERT INTO messages (id, conversation_id, sender_id, timestamp, text, photo, forwarded, reply_to)
		VALUES (NULL, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, NULLIF(?, 0))`,
		convId, senderId, timestamp, text, photo, forwarded, replyTo)
	if err != nil {
		return Message{}, err
	}

	msgId, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}

	// one status row for each other member of the conversation
	_, err = db.c.Exec(`
		INSERT INTO message_status (user_id, message_id, status)
		SELECT user_id, ?, 0 FROM members WHERE conversation_id = ? AND user_id != ?`,
		msgId, convId, senderId)
	if err != nil {
		return Message{}, err
	}

	msg := Message{
		Id:        msgId,
		Sender:    User{Id: senderId},
		Timestamp: timestamp,
		Text:      text,
		Photo:     photo,
		Forwarded: forwarded,
		ReplyTo:   replyTo,
	}
	return msg, nil
}

// MessageExists checks if the message exists in the conversation
func (db *appdbimpl) MessageExists(msgId int64, convId int64) (bool, error) {
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = ? AND conversation_id = ?)", msgId, convId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetMessage returns text and photo of a message of the conversation
func (db *appdbimpl) GetMessage(msgId int64, convId int64) (Message, error) {
	var msg Message
	err := db.c.QueryRow("SELECT id, COALESCE(text, ''), COALESCE(photo, '') FROM messages WHERE id = ? AND conversation_id = ?",
		msgId, convId).Scan(&msg.Id, &msg.Text, &msg.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return msg, ErrMessageNotFound
		}
		return msg, err
	}
	return msg, nil
}

// GetMessages returns the messages of the conversation (newest first) with their sender and status
func (db *appdbimpl) GetMessages(convId int64) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT m.id, u.id, u.name, m.timestamp,
			COALESCE(m.text, ''), COALESCE(m.photo, ''), m.forwarded, COALESCE(m.reply_to, 0),
			COALESCE((SELECT MIN(s.status) FROM message_status s WHERE s.message_id = m.id), 0)
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp DESC`, convId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var msg Message
		err = rows.Scan(&msg.Id, &msg.Sender.Id, &msg.Sender.Name, &msg.Timestamp,
			&msg.Text, &msg.Photo, &msg.Forwarded, &msg.ReplyTo, &msg.Status)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// MarkAsReceived sets status 1 (received) on all the user's statuses still at 0 (sent)
func (db *appdbimpl) MarkAsReceived(userId int64) error {
	_, err := db.c.Exec("UPDATE message_status SET status = 1 WHERE user_id = ? AND status = 0", userId)
	return err
}

// MarkAsRead sets status 2 (read) on the user's statuses in the conversation
func (db *appdbimpl) MarkAsRead(convId int64, userId int64) error {
	_, err := db.c.Exec(`
		UPDATE message_status SET status = 2
		WHERE user_id = ? AND status < 2 AND message_id IN (SELECT id FROM messages WHERE conversation_id = ?)`,
		userId, convId)
	return err
}

// GetMessageSender returns the id of the user who sent the message
func (db *appdbimpl) GetMessageSender(msgId int64, convId int64) (int64, error) {
	var senderId int64
	err := db.c.QueryRow("SELECT sender_id FROM messages WHERE id = ? AND conversation_id = ?", msgId, convId).Scan(&senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrMessageNotFound
		}
		return 0, err
	}
	return senderId, nil
}

// DeleteMessage deletes a message (status and comments are deleted by cascade)
func (db *appdbimpl) DeleteMessage(msgId int64) error {
	_, err := db.c.Exec("DELETE FROM messages WHERE id = ?", msgId)
	return err
}

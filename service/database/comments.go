package database

import (
	"database/sql"
	"errors"
)

// CommentExists checks if the user already reacted to the message
func (db *appdbimpl) CommentExists(msgId int64, userId int64) (bool, error) {
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE message_id = ? AND user_id = ?)", msgId, userId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CreateComment saves the user's reaction on the message
func (db *appdbimpl) CreateComment(msgId int64, userId int64, emoji string) (int64, error) {
	res, err := db.c.Exec("INSERT INTO comments (id, message_id, user_id, emoji) VALUES (NULL, ?, ?, ?)", msgId, userId, emoji)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetCommentSender returns the id of the user who posted the reaction
func (db *appdbimpl) GetCommentSender(commentId int64, msgId int64) (int64, error) {
	var userId int64
	err := db.c.QueryRow("SELECT user_id FROM comments WHERE id = ? AND message_id = ?", commentId, msgId).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrCommentNotFound
		}
		return 0, err
	}
	return userId, nil
}

// DeleteComment deletes a reaction
func (db *appdbimpl) DeleteComment(commentId int64) error {
	_, err := db.c.Exec("DELETE FROM comments WHERE id = ?", commentId)
	return err
}

// GetComments returns the reactions on the messages of the conversation
func (db *appdbimpl) GetComments(convId int64) ([]Comment, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.message_id, u.id, u.name, c.emoji
		FROM comments c
		JOIN messages m ON m.id = c.message_id
		JOIN users u ON u.id = c.user_id
		WHERE m.conversation_id = ?`, convId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.Id, &c.MessageId, &c.Sender.Id, &c.Sender.Name, &c.Emoji)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return comments, nil
}

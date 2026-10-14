package api

import (
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
)

// Structs sent as JSON in the responses

type User struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Photo string `json:"photo,omitempty"`
}

func (u *User) FromDatabase(dbUser database.User) {
	u.Id = dbUser.Id
	u.Name = dbUser.Name
	u.Photo = dbUser.Photo
}

type LastMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
	IsPhoto   bool      `json:"isPhoto"`
}

type ConversationSummary struct {
	Id          int64        `json:"id"`
	IsGroup     bool         `json:"isGroup"`
	Name        string       `json:"name"`
	Photo       string       `json:"photo,omitempty"`
	LastMessage *LastMessage `json:"lastMessage,omitempty"`
}

func (c *ConversationSummary) FromDatabase(dbConv database.ConversationSummary) {
	c.Id = dbConv.Id
	c.IsGroup = dbConv.IsGroup
	c.Name = dbConv.Name
	c.Photo = dbConv.Photo
	if dbConv.HasLastMessage {
		preview := []rune(dbConv.LastText)
		if len(preview) > 50 {
			preview = preview[:50]
		}
		c.LastMessage = &LastMessage{
			Timestamp: dbConv.LastTimestamp,
			Preview:   string(preview),
			IsPhoto:   dbConv.LastIsPhoto,
		}
	}
}

type Conversation struct {
	Id       int64     `json:"id"`
	IsGroup  bool      `json:"isGroup"`
	Name     string    `json:"name"`
	Photo    string    `json:"photo,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

func (c *Conversation) FromDatabase(dbConv database.ConversationSummary) {
	c.Id = dbConv.Id
	c.IsGroup = dbConv.IsGroup
	c.Name = dbConv.Name
	c.Photo = dbConv.Photo
}

type Message struct {
	Id        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Sender    User      `json:"sender"`
	Text      string    `json:"text,omitempty"`
	Photo     string    `json:"photo,omitempty"`
	Forwarded bool      `json:"forwarded"`
	ReplyTo   int64     `json:"replyTo,omitempty"`
	Status    string    `json:"status"` // sent, received or read
	Comments  []Comment `json:"comments,omitempty"`
}

func (m *Message) FromDatabase(dbMsg database.Message) {
	m.Id = dbMsg.Id
	m.Timestamp = dbMsg.Timestamp
	m.Sender.FromDatabase(dbMsg.Sender)
	m.Text = dbMsg.Text
	m.Photo = dbMsg.Photo
	m.Forwarded = dbMsg.Forwarded
	m.ReplyTo = dbMsg.ReplyTo
	m.Status = "sent"
	if dbMsg.Status == 1 {
		m.Status = "received"
	}
	if dbMsg.Status == 2 {
		m.Status = "read"
	}
}

type Comment struct {
	Id     int64  `json:"id"`
	Sender User   `json:"sender"`
	Emoji  string `json:"emoji"`
}

func (c *Comment) FromDatabase(dbComment database.Comment) {
	c.Id = dbComment.Id
	c.Sender.FromDatabase(dbComment.Sender)
	c.Emoji = dbComment.Emoji
}

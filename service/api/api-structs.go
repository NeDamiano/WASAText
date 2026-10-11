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
}

type Conversation struct {
	Id          int64        `json:"id"`
	IsGroup     bool         `json:"isGroup"`
	Name        string       `json:"name"`
	Photo       string       `json:"photo,omitempty"`
	Members     []User       `json:"members"`
	LastMessage *LastMessage `json:"lastMessage,omitempty"`
	Messages    []Message    `json:"messages,omitempty"`
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

type Comment struct {
	Id     int64  `json:"id"`
	Sender User   `json:"sender"`
	Emoji  string `json:"emoji"`
}

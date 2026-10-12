package database

import (
	"time"
)

// Structs used in the database

type User struct {
	Id    int64
	Name  string
	Photo string
}

type ConversationSummary struct {
	Id             int64
	IsGroup        bool
	Name           string
	Photo          string
	HasLastMessage bool
	LastTimestamp  time.Time
	LastText       string
	LastIsPhoto    bool
}

type Message struct {
	Id        int64
	Sender    User
	Timestamp time.Time
	Text      string
	Photo     string
	Forwarded bool
	ReplyTo   int64
	Status    int // 0 sent, 1 received, 2 read
}

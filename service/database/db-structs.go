package database

type User struct {
	Id    int64
	Name  string
	Photo string
}

type ConversationSummary struct {
	Id      int64
	IsGroup bool
	Name    string
	Photo   string
}

package database

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrConversationNotFound = errors.New("conversation not found")

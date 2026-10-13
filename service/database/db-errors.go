package database

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrConversationNotFound = errors.New("conversation not found")
var ErrMessageNotFound = errors.New("message not found")
var ErrCommentNotFound = errors.New("comment not found")

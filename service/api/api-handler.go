package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	// session
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// me
	rt.router.PUT("/me/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/me/photo", rt.wrap(rt.setMyPhoto))

	// users
	rt.router.GET("/users", rt.wrap(rt.searchUsers))

	// photos
	rt.router.GET("/photos/:photoId", rt.wrap(rt.getPhoto))

	// conversations
	rt.router.POST("/conversations", rt.wrap(rt.startConversation))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversation))

	// messages
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/conversations/:conversationId/messages/:messageId/forwards", rt.wrap(rt.forwardMessage))

	// comments
	rt.router.POST("/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrap(rt.uncommentMessage))

	return rt.router
}

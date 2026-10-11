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

	return rt.router
}

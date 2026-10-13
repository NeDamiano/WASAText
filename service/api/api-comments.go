package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

type commentMessageRequest struct {
	Emoji string `json:"emoji"`
}

// commentMessage adds a reaction (an emoji) of the logged-in user to a message
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	convId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	msgId, err := strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}

	var req commentMessageRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Emoji == "" {
		sendError(w, http.StatusBadRequest, "emoji required")
		return
	}

	// the user must be a member of the conversation
	_, err = rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("commentMessage: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	exists, err := rt.db.MessageExists(msgId, convId)
	if err != nil {
		ctx.Logger.WithError(err).Error("commentMessage: cannot check the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// one reaction per user on each message
	exists, err = rt.db.CommentExists(msgId, id)
	if err != nil {
		ctx.Logger.WithError(err).Error("commentMessage: cannot check the reaction")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		sendError(w, http.StatusConflict, "you already reacted to this message")
		return
	}

	commentId, err := rt.db.CreateComment(msgId, id, req.Emoji)
	if err != nil {
		ctx.Logger.WithError(err).Error("commentMessage: cannot save the reaction")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	sender, err := rt.db.GetUserById(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("commentMessage: cannot read the sender")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	comment := Comment{
		Id:     commentId,
		Sender: User{Id: sender.Id, Name: sender.Name},
		Emoji:  req.Emoji,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(comment)
}

// uncommentMessage removes a reaction posted by the logged-in user
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	convId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	msgId, err := strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}
	commentId, err := strconv.ParseInt(ps.ByName("commentId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}

	// the user must be a member of the conversation
	_, err = rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("uncommentMessage: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	exists, err := rt.db.MessageExists(msgId, convId)
	if err != nil {
		ctx.Logger.WithError(err).Error("uncommentMessage: cannot check the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// only the author can remove the reaction
	userId, err := rt.db.GetCommentSender(commentId, msgId)
	if errors.Is(err, database.ErrCommentNotFound) {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("uncommentMessage: cannot read the reaction")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if userId != id {
		sendError(w, http.StatusForbidden, "you can only remove your own reactions")
		return
	}

	err = rt.db.DeleteComment(commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("uncommentMessage: cannot delete the reaction")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

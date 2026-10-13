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

// sendMessage sends a new message (text, photo or both) in a conversation of the logged-in user
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	convId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}

	// the user must be a member of the conversation
	_, err = rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("sendMessage: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// text and image are both optional, but at least one is needed
	text := r.FormValue("text")
	if len(text) > 10000 {
		sendError(w, http.StatusBadRequest, "text too long")
		return
	}
	data, err := readImage(r)
	hasImage := err == nil
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		sendError(w, http.StatusBadRequest, "invalid image")
		return
	}
	if text == "" && !hasImage {
		sendError(w, http.StatusBadRequest, "text or image required")
		return
	}

	// replyTo is optional: it must be a message of the same conversation
	var replyTo int64
	if r.FormValue("replyTo") != "" {
		replyTo, err = strconv.ParseInt(r.FormValue("replyTo"), 10, 64)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid replyTo")
			return
		}
		var exists bool
		exists, err = rt.db.MessageExists(replyTo, convId)
		if err != nil {
			ctx.Logger.WithError(err).Error("sendMessage: cannot check the replied message")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !exists {
			sendError(w, http.StatusBadRequest, "replied message not found")
			return
		}
	}

	// save the image in the photos table
	photo := ""
	if hasImage {
		var photoId int64
		photoId, err = rt.db.CreatePhoto(data)
		if err != nil {
			ctx.Logger.WithError(err).Error("sendMessage: cannot save the photo")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		photo = "/photos/" + strconv.FormatInt(photoId, 10)
	}

	dbMsg, err := rt.db.CreateMessage(convId, id, text, photo, replyTo, false)
	if err != nil {
		ctx.Logger.WithError(err).Error("sendMessage: cannot save the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// read the sender (the logged-in user) for the response
	sender, err := rt.db.GetUserById(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("sendMessage: cannot read the sender")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	msg := Message{}
	msg.FromDatabase(dbMsg)
	msg.Sender = User{Id: sender.Id, Name: sender.Name}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// deleteMessage deletes a message sent by the logged-in user
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	// the user must be a member of the conversation
	_, err = rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("deleteMessage: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// only the sender can delete the message
	senderId, err := rt.db.GetMessageSender(msgId, convId)
	if errors.Is(err, database.ErrMessageNotFound) {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("deleteMessage: cannot read the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if senderId != id {
		sendError(w, http.StatusForbidden, "you can only delete your own messages")
		return
	}

	err = rt.db.DeleteMessage(msgId)
	if err != nil {
		ctx.Logger.WithError(err).Error("deleteMessage: cannot delete the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type forwardMessageRequest struct {
	DestinationId int64 `json:"destinationId"`
}

// forwardMessage copies a message into another conversation of the logged-in user
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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

	var req forwardMessageRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// the user must be a member of the source conversation
	_, err = rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	original, err := rt.db.GetMessage(msgId, convId)
	if errors.Is(err, database.ErrMessageNotFound) {
		sendError(w, http.StatusNotFound, "message not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage: cannot read the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// the user must be a member of the destination conversation too
	_, err = rt.db.GetConversation(req.DestinationId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "destination conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage: cannot read the destination conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// the copy keeps text and photo, without replyTo
	dbMsg, err := rt.db.CreateMessage(req.DestinationId, id, original.Text, original.Photo, 0, true)
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage: cannot save the message")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	sender, err := rt.db.GetUserById(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("forwardMessage: cannot read the sender")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	msg := Message{}
	msg.FromDatabase(dbMsg)
	msg.Sender = User{Id: sender.Id, Name: sender.Name}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

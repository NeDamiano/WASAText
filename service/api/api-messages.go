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

	dbMsg, err := rt.db.CreateMessage(convId, id, text, photo, replyTo)
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

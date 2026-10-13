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

type startConversationRequest struct {
	UserId int64 `json:"userId"`
}

// startConversation starts a new conversation with another user
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	var req startConversationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserId == id {
		sendError(w, http.StatusBadRequest, "cant start a conversation with yourself")
		return
	}

	// check if other user exists
	other, err := rt.db.GetUserById(req.UserId)
	if errors.Is(err, database.ErrUserNotFound) {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("startConversation: database error")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// look for existing conversation, or create it
	var status = http.StatusOK
	convId, err := rt.db.GetDirectConversation(id, other.Id)
	if errors.Is(err, database.ErrConversationNotFound) {
		convId, err = rt.db.CreateDirectConversation(id, other.Id)
		status = http.StatusCreated
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("startConversation: cant get or create the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := Conversation{
		Id:      convId,
		IsGroup: false,
		Name:    other.Name,
		Photo:   other.Photo,
	}

	// responds 200 or 201 with the conversation
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

type getMyConversationsResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}

// getMyConversations returns the logged-in user conversaions
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	// downloading the list means the new messages have been received
	err := rt.db.MarkAsReceived(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("getMyConversations: cannot mark the messages as received")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	dbConvs, err := rt.db.GetMyConversations(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("getMyConversations: cannot read the conversations")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := getMyConversationsResponse{Conversations: []ConversationSummary{}}
	for _, dbConv := range dbConvs {
		conv := ConversationSummary{}
		conv.FromDatabase(dbConv)
		resp.Conversations = append(resp.Conversations, conv)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// getConversation returns one conversation of the logged-in user, with its messages
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	convId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}

	dbConv, err := rt.db.GetConversation(convId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("getConversation: cannot read the conversation")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// opening the conversation marks its messages as read
	err = rt.db.MarkAsRead(convId, id)
	if err != nil {
		ctx.Logger.WithError(err).Error("getConversation: cannot mark the messages as read")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	dbMsgs, err := rt.db.GetMessages(convId)
	if err != nil {
		ctx.Logger.WithError(err).Error("getConversation: cannot read the messages")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	dbComments, err := rt.db.GetComments(convId)
	if err != nil {
		ctx.Logger.WithError(err).Error("getConversation: cannot read the reactions")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	summary := ConversationSummary{}
	summary.FromDatabase(dbConv)

	resp := Conversation{
		Id:       summary.Id,
		IsGroup:  summary.IsGroup,
		Name:     summary.Name,
		Photo:    summary.Photo,
		Messages: []Message{},
	}

	// convert each message and attach its reactions
	for _, dbMsg := range dbMsgs {
		msg := Message{}
		msg.FromDatabase(dbMsg)
		// the reactions of the whole conversation are read once, keep only this message's ones
		for _, dbComment := range dbComments {
			if dbComment.MessageId == msg.Id {
				comment := Comment{}
				comment.FromDatabase(dbComment)
				msg.Comments = append(msg.Comments, comment)
			}
		}
		resp.Messages = append(resp.Messages, msg)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

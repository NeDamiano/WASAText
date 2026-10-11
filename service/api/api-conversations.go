package api

import (
	"encoding/json"
	"errors"
	"net/http"

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

	// read the logged-in user (for the members list)
	me, err := rt.db.GetUserById(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("startConversation: cant read the logged-in user")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var meUser, otherUser User
	meUser.FromDatabase(me)
	otherUser.FromDatabase(other)

	resp := Conversation{
		Id:      convId,
		Type:    "direct",
		Name:    other.Name,
		Photo:   other.Photo,
		Members: []User{meUser, otherUser},
	}

	// responds 200 or 201 with the conversation
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// getMyConversationsResponse returns the logged-in user conversaions
type getMyConversationsResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
}

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
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

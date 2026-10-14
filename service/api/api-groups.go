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

type createGroupRequest struct {
	Name    string  `json:"name"`
	Members []int64 `json:"members"`
}

// createGroup creates a group with the logged-in user and the given members
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	var req createGroupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Name == "" || len(req.Members) == 0 {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// every member must be an existing user
	for _, memberId := range req.Members {
		_, err = rt.db.GetUserById(memberId)
		if errors.Is(err, database.ErrUserNotFound) {
			sendError(w, http.StatusNotFound, "user not found")
			return
		}
		if err != nil {
			ctx.Logger.WithError(err).Error("createGroup: cannot read the user")
			sendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	groupId, err := rt.db.CreateGroup(req.Name, append(req.Members, id))
	if err != nil {
		ctx.Logger.WithError(err).Error("createGroup: cannot create the group")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := Conversation{Id: groupId, IsGroup: true, Name: req.Name}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

type addToGroupRequest struct {
	UserId int64 `json:"userId"`
}

// addToGroup adds a user to a group of the logged-in user
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "group not found")
		return
	}

	var req addToGroupRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.UserId == 0 {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// only members can add, and only to a group
	conv, err := rt.db.GetConversation(groupId, id)
	if errors.Is(err, database.ErrConversationNotFound) {
		sendError(w, http.StatusNotFound, "group not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("addToGroup: cannot read the group")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !conv.IsGroup {
		sendError(w, http.StatusForbidden, "not a group")
		return
	}

	_, err = rt.db.GetUserById(req.UserId)
	if errors.Is(err, database.ErrUserNotFound) {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("addToGroup: cannot read the user")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = rt.db.AddMember(groupId, req.UserId)
	if err != nil {
		ctx.Logger.WithError(err).Error("addToGroup: cannot add the member")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := Conversation{}
	resp.FromDatabase(conv)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// leaveGroup removes the logged-in user from a group
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "group not found")
		return
	}

	// only a member can leave, and only a group
	conv, err := rt.db.GetConversation(groupId, id)
	if errors.Is(err, database.ErrConversationNotFound) || (err == nil && !conv.IsGroup) {
		sendError(w, http.StatusNotFound, "group not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("leaveGroup: cannot read the group")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = rt.db.LeaveGroup(groupId, id)
	if err != nil {
		ctx.Logger.WithError(err).Error("leaveGroup: cannot leave the group")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

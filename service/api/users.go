package api

import (
	"encoding/json"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// searchUsersResponse is the reply of GET /users: {"users": [...]}
type searchUsersResponse struct {
	Users []User `json:"users"`
}

// searchUsers lists all users, or only those whose name contains ?username=
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// authenticate the user
	_, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	// text to search for in the username (empty = all users)
	text := r.URL.Query().Get("username")

	// call db query to search users
	dbUsers, err := rt.db.SearchUsers(text)
	if err != nil {
		ctx.Logger.WithError(err).Error("searchUsers: cannot search the users")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// saves the users from db into the JSON structs
	resp := searchUsersResponse{Users: []User{}}
	for _, u := range dbUsers {
		resp.Users = append(resp.Users, User{Id: u.Id, Name: u.Name, Photo: u.Photo})
	}

	// responds 200 with the list of users
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		ctx.Logger.WithError(err).Error("searchUsers: cannot write the response")
	}
}

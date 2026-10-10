package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"unicode/utf8"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

// loginRequest is the body of POST /session: {"name": "Maria"}
type loginRequest struct {
	Name string `json:"name"`
}

// loginResponse is the reply of POST /session: {"identifier": 1}
type loginResponse struct {
	Identifier int64 `json:"identifier"`
}

// doLogin logs in the user. If the user does not exist, it's created
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	length := utf8.RuneCountInString(req.Name)
	if length < 3 || length > 16 {
		sendError(w, http.StatusBadRequest, "username must be 3-16 characters")
		return
	}

	// get user id or create user
	id, err := rt.db.GetUserByName(req.Name)
	if errors.Is(err, database.ErrUserNotFound) {
		id, err = rt.db.CreateUser(req.Name)
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("doLogin: database error")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	resp := loginResponse{Identifier: id}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

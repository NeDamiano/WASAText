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

type userNameRequest struct {
	Name string `json:"name"`
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// authenticate the user
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	// read body
	var req userNameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctx.Logger.WithError(err).Warn("setMyUserName: invalid request body")
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// check name length
	length := utf8.RuneCountInString(req.Name)
	if length < 3 || length > 16 {
		sendError(w, http.StatusBadRequest, "username must be 3-16 characters")
		return
	}

	// check if name already in use
	id2, err := rt.db.GetUserByName(req.Name)
	if err == nil && id2 != id {
		sendError(w, http.StatusConflict, "username already in use")
		return
	}
	if err != nil && !errors.Is(err, database.ErrUserNotFound) {
		ctx.Logger.WithError(err).Error("setMyUserName: cannot check the username")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// overwrite the username
	err = rt.db.SetUserName(id, req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyUserName: cannot update the username")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// responds 200 with the updated user
	resp := User{Id: id, Name: req.Name}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyUserName: cannot write the response")
	}
}

// setMyPhoto changes the profile photo of the logged-in user
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// authenticate the user
	id, ok := rt.authenticate(w, r, ctx)
	if !ok {
		return
	}

	// read the "image" field of the multipart/form-data body
	data, err := readImage(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid image")
		return
	}

	// save the image as a file in the photos folder
	photoUrl, err := savePhoto(data)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyPhoto: cannot save the photo")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// save the photo URL in the database
	err = rt.db.SetUserPhoto(id, photoUrl)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyPhoto: cannot update the photo")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// read the updated user from the database
	u, err := rt.db.GetUserById(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyPhoto: cannot read the user")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// respond 200 with the updated user
	resp := User{Id: u.Id, Name: u.Name, Photo: u.Photo}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		ctx.Logger.WithError(err).Error("setMyPhoto: cannot write the response")
	}
}

package api

import (
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
)

// authenticate reads the "Authorization: Bearer <id>" header and returns the id of the user making the request
func (rt *_router) authenticate(w http.ResponseWriter, r *http.Request, ctx reqcontext.RequestContext) (int64, bool) {

	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found {
		sendError(w, http.StatusUnauthorized, "not authorized")
		return 0, false
	}

	id, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "not authorized")
		return 0, false
	}

	exists, err := rt.db.UserExists(id)
	if err != nil {
		ctx.Logger.WithError(err).Error("authenticate: cannot check the user")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return 0, false
	}
	if !exists {
		sendError(w, http.StatusUnauthorized, "not authorized")
		return 0, false
	}

	return id, true
}

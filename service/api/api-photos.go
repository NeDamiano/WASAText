package api

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
	"github.com/julienschmidt/httprouter"
)

const maxPhotoSize = 10000000

// readImage reads the "image" field of a multipart/form-data request
func readImage(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("image")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maxPhotoSize {
		return nil, errors.New("image empty or too big")
	}
	if !strings.HasPrefix(http.DetectContentType(data), "image/") {
		return nil, errors.New("file is not an image")
	}
	return data, nil
}

// getPhoto returns the image with the given photoId (no Authorization needed)
func (rt *_router) getPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// read photoId from the path
	photoId, err := strconv.ParseInt(ps.ByName("photoId"), 10, 64)
	if err != nil {
		sendError(w, http.StatusNotFound, "photo not found")
		return
	}

	// read the image from the database
	data, err := rt.db.GetPhoto(photoId)
	if errors.Is(err, database.ErrPhotoNotFound) {
		sendError(w, http.StatusNotFound, "photo not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("getPhoto: cannot read the photo")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// respond 200 with the image
	w.Header().Set("content-type", http.DetectContentType(data))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

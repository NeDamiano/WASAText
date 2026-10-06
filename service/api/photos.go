package api

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// photosDir is the folder where the photo files are saved
const photosDir = "/tmp/wasatext-photos"

const maxPhotoSize = 10000000

// readImage reads the "image" field of a multipart/form-data request.
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

// savePhoto writes the image in photosDir with a new numeric name
func savePhoto(data []byte) (string, error) {
	err := os.MkdirAll(photosDir, 0755)
	if err != nil {
		return "", err
	}
	name := strconv.FormatInt(time.Now().UnixNano(), 10)
	err = os.WriteFile(photosDir+"/"+name, data, 0644)
	if err != nil {
		return "", err
	}
	return "/photos/" + name, nil
}

// getPhoto returns the image file with the given photoId
func (rt *_router) getPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// read photoId from the path: it must be a number
	photoId, err := strconv.ParseInt(ps.ByName("photoId"), 10, 64)
	if err != nil || photoId < 1 {
		sendError(w, http.StatusNotFound, "photo not found")
		return
	}

	// read the file
	data, err := os.ReadFile(photosDir + "/" + strconv.FormatInt(photoId, 10))
	if errors.Is(err, os.ErrNotExist) {
		sendError(w, http.StatusNotFound, "photo not found")
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("getPhoto: cannot read the photo")
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// respond 200 with the image bytes
	w.Header().Set("content-type", http.DetectContentType(data))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		ctx.Logger.WithError(err).Error("getPhoto: cannot write the response")
	}
}

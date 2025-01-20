package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorw("Internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusInternalServerError, "The server encountered a problem")

}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("Bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusNotFound, "resource not found")
}
func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusConflict, "resource not found")
}

func (app *application) unAuthorizedErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusUnauthorized, "not authorized")
}
func (app *application) unauthBasicErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized basic error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(w, http.StatusUnauthorized, "not authorized")
}

func (app *application) unauthJwtErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized jwt error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(w, http.StatusUnauthorized, "not authorized")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path)
	writeJSONError(w, http.StatusForbidden, "forbidden")
}

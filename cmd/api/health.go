package main

import (
	"net/http"

	_ "github.com/supremed3v/social-media/internal/store"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}

	// app.store.Posts.Create(r.Context(), &store.Post{})
}

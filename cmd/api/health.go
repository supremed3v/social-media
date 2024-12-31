package main

import (
	"net/http"

	"github.com/supremed3v/social-media/internal/store"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))

	app.store.Posts.Create(r.Context(), &store.Post{})
}

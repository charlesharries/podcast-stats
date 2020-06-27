package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// apiOK returns a JSON response indicating that the request was successful.
func (app *application) apiOK(w http.ResponseWriter, r *http.Request) {
	ok := map[string]interface{}{
		"error":   false,
		"message": "ok",
	}

	js, err := json.Marshal(ok)
	if err != nil {
		app.serverError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// apiListen is the API-hittable endpoint for 'listening' to an episode.
func (app *application) apiListen(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	episodeID, err := strconv.Atoi(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	err = app.listens.Create(currentUser.ID, uint(episodeID))
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	app.apiOK(w, r)
}

// apiUnlisten is the API-hittable endpoint for 'unlistening' to an episode.
func (app *application) apiUnlisten(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	episodeID, err := strconv.Atoi(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	err = app.listens.Delete(currentUser.ID, uint(episodeID))
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	app.apiOK(w, r)
}

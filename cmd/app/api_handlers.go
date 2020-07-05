package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/charlesharries/podcast-stats/pkg/forms"
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

// subscribe subscribes the currently logged in user to a podcast.
// TODO(charles): Maybe refactor some of this--it all feels a bit long.
func (app *application) apiSubscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("collectionID")
	if !form.Valid() {
		app.apiServerError(w, errors.New("collection ID is required"))
		return
	}

	collectionID, err := strconv.Atoi(form.Get("collectionID"))
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	err = app.subscriptions.Create(collectionID, currentUser.ID)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	// Save the episodes of the newly subscribed podcast in the background.
	go func(collectionID int) {
		episodes, err := app.getEpisodes(collectionID)
		if err != nil {
			app.apiServerError(w, err)
			return
		}

		err = app.saveEpisodes(collectionID, episodes)
		if err != nil {
			app.apiServerError(w, err)
		}
	}(collectionID)

	app.apiOK(w, r)
}

// unsubscribe removes a user's podcast subscription.
func (app *application) apiUnsubscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("collectionID")
	if !form.Valid() {
		app.apiServerError(w, errors.New("collection ID is required"))
		return
	}

	collectionID, err := strconv.Atoi(form.Get("collectionID"))
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	err = app.subscriptions.Delete(collectionID, currentUser.ID)
	if err != nil {
		app.apiServerError(w, err)
		return
	}

	app.apiOK(w, r)
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

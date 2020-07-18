package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/forms"
)

// serverError writes a basic 500 error as a response.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// apiServerError writes a 500 error to a JSON response.
func (app *application) apiServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)

	body := map[string]interface{}{
		"error":   true,
		"message": "Server error.",
	}
	js, err := json.Marshal(body)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	http.Error(w, string(js), http.StatusInternalServerError)
}

// clientError writes an error with whatever status code we pass in.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// addDefaultData adds some data that we're going to need on all pages.
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}

	td.CurrentYear = time.Now().Year()
	td.CurrentMonth = time.Now().Month()
	td.Calendar = newCalendar(time.Now().Year(), time.Now().Month(), -2)
	td.Flash = app.session.PopString(r, "flash")
	if app.session.Exists(r, "authenticatedUser") {
		td.User = app.session.Get(r, "authenticatedUser").(TemplateUser)
	}
	td.SearchForm = forms.New(nil)

	return td
}

// render renders out a specific page.
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s doesn't exist", name))
	}

	buf := new(bytes.Buffer)

	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	buf.WriteTo(w)
}

// isAuthenticated checks if there's a valid user in our request context.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

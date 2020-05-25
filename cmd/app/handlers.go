package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charlesharries/podcast-stats/pkg/forms"
	"github.com/charlesharries/podcast-stats/pkg/models"
)

// ping just returns a 200 OK with body "OK" to show that our
// application is running.
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// home handles GET requests for the application root.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "index.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// signupPage handles GET requests for the /signup route.
func (app *application) signupPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// signup handles POST requests to the /signup route.
func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		app.render(w, r, "signup.tmpl", &templateData{
			Form: form,
		})
	}

	err = app.users.Create(form.Get("email"), form.Get("password"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "You've been signed up.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// loginPage handles GET requests for the /login route.
func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// login handles POST requests to the /login route.
func (app *application) login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	user, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or password is incorrect.")
			app.render(w, r, "login.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}

		return
	}

	// Create a template user to pass into the template
	tu := &TemplateUser{
		ID:    user.ID,
		Email: user.Email,
	}

	app.session.Put(r, "authenticatedUser", tu)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logout removes the current user's session data.
func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUser")

	app.session.Put(r, "flash", "You've been logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// search handles searching for a podcast.
func (app *application) search(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.Form)
	form.Required("s")
	if !form.Valid() {
		app.session.Put(r, "flash", "Please enter a search term.")
		app.render(w, r, "results.tmpl", nil)
		return
	}

	result, err := app.getResults(form.Get("s"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "results.tmpl", &templateData{
		Search:  form.Get("s"),
		Results: result,
	})
}

// subscribe subscribes the currently logged in user to a podcast.
func (app *application) subscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("collectionID", "collectionName")
	if !form.Valid() {
		app.session.Put(r, "flash", "Couldn't subscribe you, sorry.")
		http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
	}

	collectionID, err := strconv.Atoi(form.Get("collectionID"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.subscriptions.Create(collectionID, currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", fmt.Sprintf("You've been subscribed to %q", form.Get("collectionName")))

	http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
}

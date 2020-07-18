package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/charlesharries/podcast-stats/pkg/forms"
	"github.com/charlesharries/podcast-stats/pkg/models"
	"github.com/jinzhu/gorm"
)

// ping just returns a 200 OK with body "OK" to show that our
// application is running.
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// home handles GET requests for the application root.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)

	subscriptions, err := app.subscriptions.FindAll(currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var ss []TemplateSubscription
	var stats TemplateStats

	for _, s := range subscriptions {
		var eps []TemplateEpisode

		listens, err := app.listens.FindByPodcast(currentUser.ID, s.Podcast.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}

		for _, ep := range s.Podcast.Episodes {
			listened := false

			for _, l := range listens {
				if l.EpisodeID == ep.ID {
					listened = true
					break
				}
			}

			eps = append(eps, TemplateEpisode{
				Title:        ep.Title,
				PublishedOn:  ep.PublishedOn,
				Duration:     ep.Duration,
				Listened:     listened,
				CollectionID: s.Podcast.ID,
			})
		}

		stats.UnlistenedEps += countUnlistened(eps)
		stats.UnlistenedTime += unlistenedTime(eps)

		ss = append(ss, TemplateSubscription{
			CollectionID: s.Podcast.ID,
			Name:         s.Podcast.Name,
			Episodes:     eps,
		})
	}

	app.render(w, r, "index.tmpl", &templateData{
		Subscriptions: ss,
		Stats:         stats,
		EpisodesByDay: episodesByDay(ss),
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

	err = app.saveResults(result.Results)
	if err != nil {
		app.serverError(w, err)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var ss []TemplateSubscription
	subscriptions, err := app.subscriptions.FindAll(currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, s := range subscriptions {
		ss = append(ss, TemplateSubscription{
			CollectionID: s.Podcast.ID,
			Name:         s.Podcast.Name,
		})
	}

	app.render(w, r, "results.tmpl", &templateData{
		Search:        form.Get("s"),
		Results:       result,
		Subscriptions: ss,
	})
}

// subscribe subscribes the currently logged in user to a podcast.
// TODO(charles): Maybe refactor some of this--it all feels a bit long.
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
		return
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

	// Check if a user is already subscribed to a podcast.
	subscription, err := app.subscriptions.Find(collectionID, currentUser.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		app.serverError(w, err)
		return
	}

	if subscription.PodcastID == collectionID {
		app.session.Put(r, "flash", fmt.Sprintf("Already subscribed to %q", form.Get("collectionName")))
		http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
		return
	}

	err = app.subscriptions.Create(collectionID, currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Save the episodes of the newly subscribed podcast.
	episodes, err := app.getEpisodes(collectionID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.saveEpisodes(collectionID, episodes)
	if err != nil {
		app.serverError(w, err)
	}

	app.session.Put(r, "flash", fmt.Sprintf("You've been subscribed to %q", form.Get("collectionName")))

	http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
}

// unsubscribe removes a user's podcast subscription.
func (app *application) unsubscribe(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("collectionID")
	if !form.Valid() {
		app.session.Put(r, "flash", "Couldn't unsubscribe you, sorry.")
		http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
		return
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

	err = app.subscriptions.Delete(collectionID, currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", fmt.Sprintf("You've been unsubscribed from %q", form.Get("collectionName")))

	http.Redirect(w, r, "/search?s="+url.QueryEscape(form.Get("search")), http.StatusSeeOther)
}

// fetchEpisodes fetches the last 20 episodes of a given podcast and saves them.
func (app *application) fetchEpisodes(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("collectionID")
	if !form.Valid() {
		app.session.Put(r, "flash", "Please submit a podcast to refetch.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	collectionID, err := strconv.Atoi(form.Get("collectionID"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Save the episodes of the newly subscribed podcast.
	episodes, err := app.getEpisodes(collectionID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.saveEpisodes(collectionID, episodes)
	if err != nil {
		app.serverError(w, err)
	}

	app.session.Put(r, "flash", "Fetched new episodes.")

	http.Redirect(w, r, fmt.Sprintf("/podcasts/%d", collectionID), http.StatusSeeOther)
}

// fetchAllUserEpisodes refetches all of a user's subscriptions
func (app *application) fetchAllUserEpisodes(w http.ResponseWriter, r *http.Request) {
	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)

	subscriptions, err := app.subscriptions.FindAll(currentUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var wg sync.WaitGroup

	for _, sub := range subscriptions {
		wg.Add(1)

		go func(sub models.Subscription) {
			// Save the episodes of the newly subscribed podcast.
			episodes, err := app.getEpisodes(sub.PodcastID)
			if err != nil {
				app.serverError(w, err)
				return
			}

			err = app.saveEpisodes(sub.PodcastID, episodes)
			if err != nil {
				app.serverError(w, err)
			}

			wg.Done()
		}(sub)
	}

	wg.Wait()

	app.session.Put(r, "flash", "Fetched new episodes.")

	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}

// podcastPage renders a page for an individual podcast.
func (app *application) podcastPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":collectionID")

	collectionID, err := strconv.Atoi(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	podcast, err := app.podcasts.Find(collectionID)

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var episodeIDs []uint
	for _, ep := range podcast.Episodes {
		episodeIDs = append(episodeIDs, uint(ep.ID))
	}

	listens, err := app.listens.FindByEpisodeIDs(currentUser.ID, episodeIDs)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var episodes []TemplateEpisode
	for _, ep := range podcast.Episodes {
		listened := false

		for _, l := range listens {
			if l.EpisodeID == ep.ID {
				listened = true
				break
			}
		}

		episodes = append(episodes, TemplateEpisode{
			ID:          ep.ID,
			Title:       ep.Title,
			Duration:    ep.Duration,
			PublishedOn: ep.PublishedOn,
			Listened:    listened,
		})
	}

	app.render(w, r, "podcast.tmpl", &templateData{
		Podcast:  podcast,
		Episodes: episodes,
	})
}

// listen creates a new episode listen for the logged-in user.
func (app *application) listen(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	episodeID, err := strconv.Atoi(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.listens.Create(currentUser.ID, uint(episodeID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func (app *application) unlisten(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":id")

	episodeID, err := strconv.Atoi(id)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentUser := app.session.Get(r, "authenticatedUser").(TemplateUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.listens.Delete(currentUser.ID, uint(episodeID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

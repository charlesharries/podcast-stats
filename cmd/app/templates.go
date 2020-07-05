package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/forms"
	"github.com/charlesharries/podcast-stats/pkg/models"
)

// TemplateUser is a representation of a user as passed to a
// template. We don't want to pass passwords and stuff into
// our templates.
type TemplateUser struct {
	ID    uint
	Email string
}

// TemplateSubscription is a representation of a subscription
// passed in to a template. We only need a subset of subscription
// data in our templates.
type TemplateSubscription struct {
	CollectionID int
	Name         string
	Episodes     []TemplateEpisode
}

// TemplateEpisode is a representation of a single podcast
// episode passed into a template. We only need a subset of episode
// data in our templates.
type TemplateEpisode struct {
	ID          uint
	Title       string
	Duration    int
	PublishedOn time.Time
	Listened    bool
}

// TemplateStats are general global stats about all of your podcasts.
type TemplateStats struct {
	UnlistenedTime int
	UnlistenedEps  int
}

type templateData struct {
	CurrentYear   int
	Flash         string
	Episodes      []TemplateEpisode
	Form          *forms.Form
	Podcast       models.Podcast
	Results       ITunesResult
	Search        string
	SearchForm    *forms.Form
	Stats         TemplateStats
	Subscriptions []TemplateSubscription
	User          TemplateUser
}

// humanDate formats time.Time objects into a human-readable format.
func humanDate(t time.Time) string {
	// Return empty if the time has zero value.
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// inSlice checks if a value can be found in a slice.
func hasSubscription(ss []TemplateSubscription, id int) bool {
	for _, s := range ss {
		if s.CollectionID == id {
			return true
		}
	}

	return false
}

// countUnlistened gets the number of unlistened-to episodes.
func countUnlistened(eps []TemplateEpisode) int {
	count := 0
	for _, ep := range eps {
		if !ep.Listened {
			count++
		}
	}

	return count
}

func humanSeconds(secs int) string {
	h := secs / (60 * 60)
	m := (secs - (h * 60 * 60)) / 60

	hs := ""
	ms := ""

	if h > 0 {
		hs = fmt.Sprintf("%dh ", h)
	}

	if m > 0 {
		ms = fmt.Sprintf("%dm", m)
	}

	return hs + ms
}

// unlistenedTime get the amount of unlistened-to podcast time.
func unlistenedTime(eps []TemplateEpisode) int {
	seconds := 0

	for _, ep := range eps {
		if !ep.Listened {
			seconds += ep.Duration
		}
	}

	return seconds
}

// byPublishedOn is a custom sort type.
type byPublishedOn []TemplateEpisode

// Len returns the length of the sortable.
func (b byPublishedOn) Len() int {
	return len(b)
}

// Swap indicates how to swap two sortable elements.
func (b byPublishedOn) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less handles the actual sorting logic.
func (b byPublishedOn) Less(i, j int) bool {
	return b[i].PublishedOn.After(b[j].PublishedOn)
}

// soryByPublishedOn sorts TemplateEpisodes by PublishedOn times.
func sortByPublishedOn(eps []TemplateEpisode) []TemplateEpisode {
	sort.Sort(byPublishedOn(eps))
	return eps
}

// functions passes some functions into our templates.
var functions = template.FuncMap{
	"humanDate":         humanDate,
	"hasSubscription":   hasSubscription,
	"countUnlistened":   countUnlistened,
	"sortByPublishedOn": sortByPublishedOn,
	"unlistenedTime":    unlistenedTime,
	"humanSeconds":      humanSeconds,
}

// newTemplateCache pre-compiles all of our templates so we're not re-compiling
// on every request.
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "pages/*"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Compile all of our pages
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Compile all of our layouts
		ts, err = ts.ParseGlob(filepath.Join(dir, "layouts/*"))
		if err != nil {
			return nil, err
		}

		// Compile all of our components
		ts, err = ts.ParseGlob(filepath.Join(dir, "components/*"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

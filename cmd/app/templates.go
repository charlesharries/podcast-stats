package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/charlesharries/podcast-stats/pkg/forms"
)

// TemplateUser is a representation of a user as passed to a
// template. We don't want to pass passwords and stuff into
// our templates.
type TemplateUser struct {
	Email string
}

type templateData struct {
	CurrentYear int
	Flash       string
	Form        *forms.Form
	User        TemplateUser
}

// humanDate formats time.Time objects into a human-readable format.
func humanDate(t time.Time) string {
	// Return empty if the time has zero value.
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// functions passes some functions into our templates.
var functions = template.FuncMap{
	"humanDate": humanDate,
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

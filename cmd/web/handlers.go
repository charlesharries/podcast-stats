package main

import "net/http"

// ping just returns a 200 OK with body "OK" to show that our
// application is running.
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// home handles GET requests for the application root.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is the homepage!"))
}

package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()

	mux.Get("/", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.home)))

	// Users routes.
	mux.Get("/signup", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.signupPage)))
	mux.Post("/signup", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.signup)))
	mux.Get("/login", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.loginPage)))
	mux.Post("/login", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.login)))
	mux.Post("/logout", dynamicMiddleware.ThenFunc(http.HandlerFunc(app.logout)))

	mux.Get("/ping", http.HandlerFunc(ping))

	return standardMiddleware.Then(mux)
}

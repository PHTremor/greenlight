package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// initialize httpRouter instance
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler
	// and then set it as the custom error handler for 404 Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// do the same for methodNot Allowed Responses 405
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// register HTTP methods, URL patterns, & handler functions

	// Handlers for movies
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.requirePermission("movies:read", app.healthcheckHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	// Handlers for users
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// Handlers for tokens
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// return the httpRouter instance
	// wrap router with panic recovery, rateLimit(), & authenticate() middlewares
	// to run for every endpoint! or request
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}

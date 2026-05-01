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
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// return the httpRouter instance
	// wrap router with panic recovery middleware to run for every endpoint!
	return app.recoverPanic(router)
}

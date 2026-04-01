package main

import (
	"fmt"
	"net/http"
)

// helper for logging an error message with the current request method and url
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "url", uri)
}

// helper for sending a JSON formated Error Response with status code to the client
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	// write the response
	err := app.writeJSON(w, status, env, nil)
	// if it fails send back an empty response with 500 Internal Server Error status code
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// sends a generic message to the client when the app encounters an unxpected problem at runtime
// ... logs the error and sends a 500 Internal Server Error status code
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your  request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// sends a 404 Not Found status code and  JSON response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// send a 405 Method Not Allowed status code and JSON response to the client
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

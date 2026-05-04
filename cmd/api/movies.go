package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PHTremor/greenlight.git/internal/data"
	"github.com/PHTremor/greenlight.git/internal/validator"
)

// add a createMovieHandler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "create a movie")

	// a struct holding the information we expect in the HTTP request body
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		// use the bad request helper
		app.badRequestResponse(w, r, err)
		return
	}

	// copy input struct into the movie struct
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// initailize an instance of the validator
	v := validator.New()

	// call ValidateMovie function to perform the checks
	// 	return the failedValidation response if checks fail
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// call the Insert() method to save the movie record in the database
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// include a Location header in the response to help the clint know where to find the created record
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write a JSON response with a 201 created status code, the movie data in the response body
	// and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// add a showMovieHandler for the "GET /v1/movies/:id" endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// call the Get() method to retrieve the movie data from the database
	// use errors.Is() to check if the error returned is data.ErrRecordNotFound and send a 404 to the client
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler for the "PUT /v1/movies/:id" endpoint
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract movie ID from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// fetch the movie from the database, send 404 if none found
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// a struct to hold the expected values from the client
	// we use pointer fields to allow partial updates, if fields are nil then the client didn't provide a value
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// read the json's body into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// leave values unchanged if the client didn't provide them in the request body
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}
	// // copy values from the input struct to their respective fields in the movie record
	// movie.Title = input.Title
	// movie.Year = input.Year
	// movie.Runtime = input.Runtime
	// movie.Genres = input.Genres

	// validate the updated movie records, send a 422 Unprocessable Entity if check fails
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// pass the movie recors to the update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Write the updated movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteMovieHandler for the "DELETE /v1/movies/:id" endpoint
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the ID from the url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// delete movie from the database, return 404 not found to client if there's no match
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// return a 200 OK status along with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": "movie successfuly deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

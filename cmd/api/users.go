package main

import (
	"errors"
	"net/http"

	"github.com/PHTremor/greenlight.git/internal/data"
	"github.com/PHTremor/greenlight.git/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// create a struct to hold the expected data from the request
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// parse the request into the input struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy the input/request body into a new user struct
	// we set activated to false for readability; it is non-zero (false) by default
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// use Password.set() to generate hash pass and store both hash & plaintext pass
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// validate the user struct and return err message if anything fails
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the user into the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// if we get a ErrDuplicateEmail, manually add a message to the vaidator instance
		// and call the failedValidationResponse() helper
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// write a JSON response with the user data along with a 201 status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

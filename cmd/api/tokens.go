package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/PHTremor/greenlight.git/internal/data"
	"github.com/PHTremor/greenlight.git/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse the request from the body
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// validate the provided email and password
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// lookup the user records based on the email address
	// return unauthorized 401 if no record was found
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// check if the provided password matches the actual password
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !matches {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// if it matches, generate a token with 24hrs expiry time and scope authentication
	token, err := app.models.Token.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// encode the token to JSON ans send it in the response along with a 201 status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

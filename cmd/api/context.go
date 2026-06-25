package main

import (
	"context"
	"net/http"

	"github.com/PHTremor/greenlight.git/internal/data"
)

type contextKey string

// convert string "user" to a contextKey type
const userContextKey = contextKey("user")

// contextSetUser() method returns a new copy of the request
// with the provided User strict attached to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

// contextGetUser() retrieves the User struct from the request context
// when in use, we expect the User to be there, otherwise it's OK to panic
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in the context")
	}

	return user
}

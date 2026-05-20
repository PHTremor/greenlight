package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// will always be run in the event of a panic as Go unwinds the stack).
		defer func() {
			if err := recover(); err != nil {
				// automatically close the current connection after a response has been sent.
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// initialize a new rate limiter that allows 2 request per second
	// with a maximum of 4 request in a single burst
	limiter := rate.NewLimiter(2, 4)

	// a closure function which closes pver the limiter variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// limiter.Allow() checks if the request is allowed, if its not
		// return a 429 Too Many Requests response
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

}

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"

	"github.com/PHTremor/greenlight.git/internal/data"
	"github.com/PHTremor/greenlight.git/internal/validator"
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
	// a client struct to hold the rate limiter and lastSeen time for each client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// create a mutex and a map to hold the clients IP addreses and RateLimiters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// a background goroutine which removed old entries from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock the mutex to prevent any rate limiter from making checks
			// while the cleanup is in process
			mu.Lock()

			// loop through clients and delete any entries that havent been seen in the last 3 minutes
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// unlock the mutex after the cleanup
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// use realip.fromRequest() to get the client's IP address.
			ip := realip.FromRequest(r)

			// lock the mutex to prevent this code from executing concurrently
			mu.Lock()

			// check if the IP address already exists in the map. If it doesn't
			// initialize a new rateLimiter, add the IP and the rateLimiter to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			// update the lastSeen time of the client
			clients[ip].lastSeen = time.Now()

			// call the Allow() method on the limiter for the current IP Address
			// If request is not allowed, unlock the mutex & send a 429 Too Many Request response
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// unlock the mutex before calling the next handler in the chain
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})

	// NOTE: Global Rate Limiter impl.....
	// // initialize a new rate limiter that allows 2 request per second
	// // with a maximum of 4 request in a single burst
	// limiter := rate.NewLimiter(2, 4)

	// // a closure function which closes pver the limiter variable
	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	// limiter.Allow() checks if the request is allowed, if its not
	// 	// return a 429 Too Many Requests response
	// 	if !limiter.Allow() {
	// 		app.rateLimitExceededResponse(w, r)
	// 		return
	// 	}

	// 	next.ServeHTTP(w, r)
	// })

}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add the "Vary: authorization" header to the response
		// it indicates to caches that the response might vary based on the value
		// of the Authorization header in the request
		w.Header().Add("Vary", "Authorization")

		// retrieve the value of the Authorization header from the request
		// it returns an empty string "" if not found
		authorizationHeader := r.Header.Get("Authorization")

		// if there's not Authorization header, use contextSetUser()
		// to add an anonymous user to the request. Then call the next Handler in the chain & return
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// otherwise we expect an Authorization token in the format "Bearer <token>"
		// if wrong format, return 401 Unauthorized response
		headerparts := strings.Split(authorizationHeader, " ")
		if len(headerparts) != 2 || headerparts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// extract the token from the header
		token := headerparts[1]

		// validate the token
		v := validator.New()

		// if token is invalid
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// retrieve the user details associated with the Authentication token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// add the user details to the request context
		r = app.contextSetUser(r, user)

		// call the next Handler in the chain
		next.ServeHTTP(w, r)
	})
}

// checks if the user is not anonymous (authentucated)
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve user details from the request context
		user := app.contextGetUser(r)

		// if user is anonymous inform the client to authenticate
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checks if the user is both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve user details from the request context
		user := app.contextGetUser(r)

		// if user is not activated, inform the client to activate
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// else, call the next handler in the chain
		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

// checks if the user has required permissions, (and is activated & authenticated)
//
// it wraps requireActivatedUser() & requireAuthenticatedUser() middlewares to check if
// user is activated & authenticated
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fun := func(w http.ResponseWriter, r *http.Request) {
		// rettrive the user from the request context
		user := app.contextGetUser(r)

		// get the slice of permissions of the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// check if the slice include the required permission
		// return a 403 if not
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// otherwise call the next handler in the chain
		next.ServeHTTP(w, r)
	}

	// wrap this middleware with the requireActivatedUser() middleware before returning it
	return app.requireActivatedUser(fun)
}

// enables CORS
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add the vary headers
		w.Header().Add("Vary", "Origin")
		// for preflight CORS
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// get the value of the request header
		origin := r.Header.Get("Origin")

		// only run this if there's an Origin request header
		if origin != "" {
			// loop through the list of trusted origins and check if any match with the request origin
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					// set the response header
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// if request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header, treat it as a preflight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// set the necessary preflight headers
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// write the headers along  with a 200 OK status and return from the middleware
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

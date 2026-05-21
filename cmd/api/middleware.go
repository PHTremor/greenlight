package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
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

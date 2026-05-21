package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// declare HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// shutdownError channel receives errors returned by the graceful Shutdown() function
	shutdownError := make(chan error)

	// start a background goroutine
	go func() {
		// create a channel which carries os.signal values
		quit := make(chan os.Signal, 1)

		// listen for incoming SIGINT & SIGTERM signals and relay them to the quit channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// read a signal from the channel
		s := <-quit

		// log a message to show that a signal has been caught
		app.logger.Info("shutting down server", "signal", s.String())

		// create a context with a 30-second timeout
		ctx, cancle := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancle()

		// call shutdown() on the server; returns nil if graceful shutdown is successful
		// or an error. relay the return value to the shutdownError channel
		shutdownError <- srv.Shutdown(ctx)

	}()

	// start the HTTP Server
	app.logger.Info("Starting server", "addr", srv.Addr, "env", app.config.env)

	// calling shutdown() on the server causes ListenAndServe to return a http.ErrServerClosed err
	// this err indicates the graceful shutdown has started. We'll only return if it's a different err
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// else, wait to recieve a value from shutdown() on the shutdownError channel
	// if the return value is an error, return the err
	err = <-shutdownError
	if err != nil {
		return err
	}

	// at this point the graceful shutdown completed successfully
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// config settings
type config struct {
	port int
	env  string
}

// dependencies (Helpers, Handlers, Middleware)
type application struct {
	config config
	logger *slog.Logger
}

func main() {
	// config instance declaration
	var cfg config

	// read env & port values from cmd flags
	// default to 4000 & development if no flags provided
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// initialise Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// declare instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// start the HTTP Server
	logger.Info("Starting server", "addr", srv.Addr, "env", cfg.env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/PHTremor/greenlight.git/internal/data"
)

const version = "1.0.0"

// config settings
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

// dependencies (Helpers, Handlers, Middleware)
type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	// config instance declaration
	var cfg config

	// read env & port values from cmd flags
	// default to 4000 & development if no flags provided
	flag.IntVar(&cfg.port, "port", 4000, "API Server Port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// read the dsn, default to the dev dsn if not provided
	// postgres://user:password@host/dbname
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// read the db connection pool settings from cmd flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Parse()

	// initialise Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// call openDB() function to create a connection pool & pass it int the config struct
	// if it errors, log it & exit the app
	db, err := openDB(&cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// set the max number of open (idle + in-use) connections in the pool
	// values less than or equal to 0 means there's no limit
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// set the max number of idle connections in the pool
	// values less than or equal to 0 means there's no limit
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// set the max idle time for connections in the pool
	// less than || equal to 0 means connections will not be closed due to their idle time
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// defer a call to db.Close() so that the connection pool is closed before the main function exits
	defer db.Close()

	// log message for successful DB conn
	logger.Info("db connection pool established")

	// declare instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
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

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB() helper method returns an sql.db connection pool
func openDB(cfg *config) (*sql.DB, error) {
	// create an empty connection pool using the dsn from cfg
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// create a context with a 5-second timeout deadline
	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	// establish new connection to the DB
	// if it fails to connect within the 5 second deadline, close the connection pool &return an error
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the db connection pool
	return db, nil
}

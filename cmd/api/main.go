package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/PHTremor/greenlight.git/internal/data"
	"github.com/PHTremor/greenlight.git/internal/mailer"
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
	// rate limiter config
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// dependencies (Helpers, Handlers, Middleware)
type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer *mailer.Mailer
	wg     sync.WaitGroup
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
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// read the rate limiter settings from cmd flags
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// read the smtp server settings from cmd flags
	// the default settings are from mailtrap
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "af9e1ded4edd88", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "ead61f83702b1d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.frankmwale.net>", "SMTP sender")

	// process the -cors-trusted-origind command line flag
	// split the flag value into a slice based on whitespace characters
	flag.Func("cors-trusted-origins", "Trusted CORS origns (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

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

	// initialize a new mailer instance
	mailer, err := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// publish a new "version" variable in the expvar Handler holding the app's version number
	expvar.NewString("version").Set(version)

	// publish the number of active goroutines
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// publish the database pool statistics
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// publish the current unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	// declare instance of the application struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}
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

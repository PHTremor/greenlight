package data

import (
	"database/sql"
	"errors"
)

// define a a custom ErrRecordNotFound error
// we'll return this in our Get() method if a movie doesnt exist
var (
	ErrRecordNotFound = errors.New("record not found")
)

// a Models struct that wraps other model
type Models struct {
	Movies MovieModel
}

// New() method that returns a Models struct with initialized models
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}

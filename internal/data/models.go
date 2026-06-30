package data

import (
	"database/sql"
	"errors"
)

// define a a custom ErrRecordNotFound error
// we'll return this in our Get() method if a movie doesnt exist
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// a Models struct that wraps other model
type Models struct {
	Movies      MovieModel
	Permissions PermissionModel
	Token       TokenModel
	Users       UserModel
}

// New() method that returns a Models struct with initialized model instances
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Token:       TokenModel{DB: db},
		Users:       UserModel{DB: db},
	}
}

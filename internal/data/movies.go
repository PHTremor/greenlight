package data

import (
	"time"

	"github.com/PHTremor/greenlight.git/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`              // - hides the field in the JSON response
	Title     string    `json:"title,omitzero"` // omitzero hides the field if it has a 0 value (false/0/nil/"")
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // string forces output field to be a string
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// use the Check method to do the validation checks

	// validate Title
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more that 500 bytes long")

	// validate year
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// validate runtime
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a postive integer")

	// validate genre
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) >= 1, "genres", "must not contain more that 5 genre")
	// use the Unique helper to make sure the genres are unique
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

package data

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

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

// Define a MovieModel stuct that wraps a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

// Inserting a new recors in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	// SQL query for inserting a new record & returning system generated data
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
	`

	// create an arg slice containing the values for the placeholder parameters in the SQL query
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use QueryRow() method to execute the query on the connection pool
	// Scan the system generated values into the Movie struct
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Fetching a specific record in the movies table
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Updating a specific record in the movies table
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Deleting a specific record in the movies table
func (m MovieModel) Delete(id int64) error {
	return nil
}

package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`              // - hides the field in the JSON response
	Title     string    `json:"title,omitzero"` // omitzero hides the field if it has a 0 value (false/0/nil/"")
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // string forces output field to be a string
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}

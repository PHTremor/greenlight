package data

import (
	"strings"

	"github.com/PHTremor/greenlight.git/internal/validator"
)

// Filters type to hold the query string values for pagination and sorting
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SafeSortlist []string
}

// Metadata struct for holding the pagination metadata values
type Metadata struct {
	CurrentPage  int `json:"current_page,omitzero"`
	PageSize     int `json:"page_size,omitzero"`
	FirstPage    int `json:"first_page,omitzero"`
	LastPage     int `json:"last_page,omitzero"`
	TotalRecords int `json:"total_records,omitzero"`
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// check page and page_size paramaters contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// check the sort parameter matches a value in the safelist
	v.Check(validator.PermittedValue(f.Sort, f.SafeSortlist...), "sort", "invalid sort value")
}

// check if the client's sort fields match our safeList
// extract the column name and stripe away the leading hyphen if it exixts
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SafeSortlist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

// return the sort direction (ASC, DESC) depending on the prefix of the sort value
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// return the limit and offset values for the SQL query based on the current page and page size
func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// calculate pagination metadata values given
// the total number of records, current page, and page size values
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// return an empty Metadata struct if there are no records
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

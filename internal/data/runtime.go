package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// error returned when UnMarshalJSON method fails to parse/convert the JSON string
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Runtime is a custom type which represents the runtime of a movie in minutes
type Runtime int32

// Return the JSON-encoded value for the movie runtime
// (it will return a string in the format "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJsonValue := strconv.Quote(jsonValue)

	return []byte(quotedJsonValue), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJsonValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// split the string to isolate the number
	parts := strings.Split(unquotedJsonValue, " ")

	// check the parts to ensure they satisfy the format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// parse the string containing the number to int32
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// converts the int32 to a Runtime type and assign it to the receiver
	// we use a pointer to set the underlying value & not it's copy, respecting the receiver
	*r = Runtime(i)

	return nil
}

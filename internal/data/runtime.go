package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// Return the JSON-encoded value for the movie runtime
// (it will return a string in the format "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJsonValue := strconv.Quote(jsonValue)

	return []byte(quotedJsonValue), nil
}

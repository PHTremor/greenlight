package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/PHTremor/greenlight.git/internal/validator"
)

// return id from from current request context
// convert it to int64 and return... or retun 0 and an error
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// retrieve parameter names and values
	params := httprouter.ParamsFromContext(r.Context())

	// get value of the id parameter
	// ByName returns a string, we'll convert it to a base 10 int (with a bit size of 64)
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		// id is invalid
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]any

// Helper for sending JSON Responses
// It expects the destination's http.ResponseWriter, the status-code to send
// the data to encode, & additional map of http headers
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// encode data to json
	// json.Marshal(data) is straight forward
	// json.MarshalIndent() add white spaces to make them readable on terminals
	// but there's a performance trade-off (milliseconds)
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// append new line to make it readable on terminal
	js = append(js, '\n')

	// add headers to the ResponseWriter header map
	for key, value := range headers {
		w.Header()[key] = value
	}

	// code below does the same work as above: chose one..lol
	// maps.Insert(w.Header(), maps.All(headers))

	// write content-type, status, and the json response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// limit the size of the request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	// return an error for any fields that can not be mapped to the target destination
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	//  Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// err has type syntaxError
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)

		// decode returns io.ErrUnexpectedEOF for syntax errors in JSON
		// return a generic message
		// follow the issue: https://github.com/golang/go/issues/25956
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")

		// unmarshalTypeError happens when the JSON value is the wrong type for the target destination
		// includes a specific field if the error relates to one
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type on field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// Decode returns io.EOF when the request body is empty
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty")

		// Decode() will now return an error message in the format "json: unknown ield "<name>"".
		// extract the field name & interporate it in our custome error
		// there's an open issue at https://github.com/golang/go/issues/29035
		// regarding turning this into a distinct error type
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldname := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldname)

		// check maxByteErrors, body shouldn't exceed 1MB
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// invalidUnmarshalError is returned when we pass a non-nil pointer to Decode
		// This doesnt need to happen, so we panic
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// for anything else, return as is
		default:
			return err
		}
	}

	// call Decode() using a pointer to an anonymous struct at the destination
	// a single JSON value will return an io.EOF error
	// anything else means there's additional data in the request body
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString() returns a value from the query string
// or the provided default value of no matching key was found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract value from query string of the given key
	// returns empty if key doesn't exist
	s := qs.Get(key)

	// return default if value is empty
	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV reads a string value from the query string and splist it into a slice on the comma character
// returns default value if no matching key was found
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {

	// extract the value from the query string
	csv := qs.Get(key)

	// return default if no key or extract is ""
	if csv == "" {
		return defaultValue
	}

	// otherwise parse the value into a string slice and return it
	return strings.Split(csv, ",")
}

// readString() reads a string value and converts it into an Integer
// returns defaultVaue If no matching key was found
// records an error in the provided validator instance if conversion fails
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// read the value
	s := qs.Get(key)

	// return default key if key doesn't exist or value is empty
	if s == "" {
		return defaultValue
	}

	// convert value to int or return an err
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

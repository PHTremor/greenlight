package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
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

package io

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Request represents a request with a generic data type to be unmarshalled
// from the request body. It implements the Read method to read and
// unmarshal the request body into the Data field.
type Request[T any] struct {
	Data        T
	initialized bool
}

// RequestWith creates a new Request[T] with the provided data. If the data
// is nil, it returns an empty Request[T], if not, the data is wrapped and
// the resulting request is marked as initialized. This method is useful for
// creating requests with different data types without needing to define
// separate request structs for each type.
func RequestWith[T any](data *T) *Request[T] {
	if data == nil {
		return &Request[T]{}
	}
	return &Request[T]{
		Data:        *data,
		initialized: true,
	}
}

// Read reads the request body and unmarshals it into the Data field of the
// Request struct. It returns an error if the request body is nil or empty,
// or if there is an error during unmarshalling. If the Request struct is nil,
// it initializes a new instance of Request[T]. This method is useful for
// handling incoming requests in a generic way, allowing for different data
// types to be processed without needing to define separate request structs
// for each type.
func (req *Request[T]) Read(r *http.Request) error {
	if req == nil {
		req = new(Request[T])
	}
	if r == nil {
		return ErrNilRequest
	}
	if r.Body == nil {
		return ErrNilRequestBody
	}
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return ErrReadBody
	}
	if len(rawBody) == 0 {
		return ErrEmptyRequestBody
	}
	if err := json.Unmarshal(rawBody, &req.Data); err != nil {
		return ErrDecodeBody
	}
	return nil
}

// WriteJSON writes the request data to the provided http.Request in JSON
// format. It sets the Content-Type header to "application/json" and replaces
// the request body with the JSON marshalled data. If the request is not
// initialized, it returns an error. This method is useful for sending
// requests with structured JSON data. It can be used in HTTP handlers
// or middleware to send requests with specific data types.
func (req *Request[T]) WriteJSON(r *http.Request) error {
	if r == nil {
		return ErrNilRequest
	}
	if req == nil || !req.initialized {
		return ErrNotInitializedRequest
	}
	// set the content type to application/json
	r.Header.Set("Content-Type", "application/json")
	// replace the request with a new one with the same method and URL
	// but with the request body set to the JSON marshalled data
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(req.Data); err != nil {
		return ErrEncodeBody
	}
	r.Body = io.NopCloser(buf)
	r.ContentLength = int64(buf.Len())
	return nil
}

package io

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response represents a response with a generic data type. It can be used to
// send JSON responses or plain text responses. The Data field holds the
// response data, and the empty field indicates whether the response is empty
// or not. It has methods to write the response to an http.ResponseWriter in
// JSON format or plain text format.
type Response[T any] struct {
	Data  T
	empty bool
}

// ResponseWith creates a new Response instance with the provided data. If
// the data is nil, it returns an empty response. This method is useful for
// creating responses with different data types without needing to define
// separate response structs for each type. It returns a pointer to the
// Response instance.
func ResponseWith[T any](data *T) *Response[T] {
	if data == nil {
		return &Response[T]{empty: true}
	}
	return &Response[T]{
		Data:  *data,
		empty: false,
	}
}

// OkResponse creates a new Response instance with the provided byte slice.
// If the byte slice is empty, it returns an empty response. This method is
// useful for creating responses with raw byte data. It returns a pointer to
// the Response instance. The empty field indicates whether the response is
// empty or not. If the byte slice is empty, the response is considered empty.
// If the byte slice is not empty, the response is considered non-empty.
func OkResponse(body ...byte) *Response[any] {
	if len(body) > 0 {
		return &Response[any]{Data: body, empty: false}
	}
	return &Response[any]{empty: true}
}

// WriteJSON writes the response data to the provided http.ResponseWriter in
// JSON format. It sets the Content-Type header to "application/json" and
// writes the response data as JSON. If the response is empty, it writes a
// plain text "OK" response with a 200 OK status code. If there is an error
// during JSON encoding or response writing, it writes an error response with
// a 500 Internal Server Error status code. This method is useful for sending
// JSON responses to the client. It can be used in HTTP handlers or middleware
// to send structured JSON responses.
func (r *Response[T]) WriteJSON(w http.ResponseWriter) {
	if !r.empty {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(r.Data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Write writes the response data to the provided http.ResponseWriter in
// plain text format. It sets the Content-Type header to "text/plain" and
// writes the response data as plain text. If the response is empty, it writes
// a plain text "OK" response with a 200 OK status code. If there is an error
// during response writing, it writes an error response with a 500 Internal
// Server Error status code. This method is useful for sending plain text
// responses to the client. It can be used in HTTP handlers or middleware to
// send simple text responses. It is a more generic method than WriteJSON, as
// it does not require the response data to be JSON-serializable.
func (r *Response[T]) Write(w http.ResponseWriter) {
	if r.empty {
		if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if data, ok := r.bytes(); ok {
		if _, err := w.Write(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Read reads the response data from the provided http.Response. It expects
// the response body to be in JSON format and unmarshals it into the provided
// generic type T. If the response status code is not 200 OK, it returns an
// error indicating the unexpected status code. If there is an error reading
// the response body or unmarshaling the JSON data, it returns an error. This
// method is useful for handling HTTP responses in a generic way, allowing
// you to read and process the response data without needing to know the
// specific type of the response data in advance.
func (r *Response[T]) Read(res *http.Response) (T, error) {
	if res == nil {
		return *new(T), ErrNilRequest
	}
	if res.Body == nil {
		return *new(T), ErrNilRequestBody
	}
	defer res.Body.Close()
	result := new(T)
	// read the response body
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return *result, ErrReadBody
	}
	// check the status code of the response
	if res.StatusCode != http.StatusOK {
		return *result, ErrHTTPStatusNotOk
	}
	// unmarshal the response body into the result
	if err := json.Unmarshal(rawBody, result); err != nil {
		return *result, ErrDecodeBody
	}
	return *result, nil
}

// bytes returns the response data as a byte slice. If the response is empty,
// it returns nil and a boolean indicating that the response is empty. If the
// response data is not a byte slice, it returns nil and a boolean indicating
// that the response data is not a byte slice. This method is useful for
// converting the response data to a byte slice for writing to the response
// writer or for further processing.
func (r *Response[T]) bytes() ([]byte, bool) {
	// check if the response is empty
	if r.empty {
		return nil, true
	}
	// ensure that the response data is an slice of bytes
	switch v := any(r.Data).(type) {
	case []byte:
		return v, true
	case string:
		return []byte(v), true
	default:
		return nil, false
	}
}

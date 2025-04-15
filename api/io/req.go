package io

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Request represents a request with a generic data type to be unmarshalled
// from the request body. It implements the Read method to read and
// unmarshal the request body into the Data field.
type Request[T any] struct {
	Data T
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
	if r.Body == nil {
		return fmt.Errorf("nil request body")
	}
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if len(rawBody) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(rawBody, &req.Data)
}

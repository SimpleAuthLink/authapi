package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Response[T any] struct {
	Data  T
	empty bool
}

func ResponseWith[T any](data *T) *Response[T] {
	if data == nil {
		return &Response[T]{empty: true}
	}
	return &Response[T]{
		Data:  *data,
		empty: false,
	}
}

func OkResponse() *Response[any] {
	return &Response[any]{empty: true}
}

func (r *Response[T]) Write(w http.ResponseWriter) error {
	if !r.empty {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(r.Data)
	}
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	return err
}

type TokenRequest struct {
	Email string `json:"email"`
}

type Request[T any] struct {
	Data T
}

func (req *Request[T]) Read(r *http.Request) error {
	if req == nil {
		req = new(Request[T])
	}
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(rawBody) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(rawBody, &req.Data)
}

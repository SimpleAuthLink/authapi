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

func (r *Response[T]) WriteJSON(w http.ResponseWriter) {
	if !r.empty {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(r.Data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type Request[T any] struct {
	Data T
}

func (req *Request[T]) Read(r *http.Request) error {
	if req == nil {
		req = new(Request[T])
	}
	if r.Body == nil {
		return fmt.Errorf("nil request body")
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

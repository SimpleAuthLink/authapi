package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func OkResponse(body ...byte) *Response[any] {
	if len(body) > 0 {
		return &Response[any]{Data: body, empty: false}
	}
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

func (r *Response[T]) Write(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
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

func (r *Response[T]) bytes() ([]byte, bool) {
	// check if the response is empty
	if r.empty {
		return nil, true
	}
	// ensure that the response data is an slice of bytes
	switch v := any(r.Data).(type) {
	case []byte:
		return v, true
	default:
		return nil, false
	}
}

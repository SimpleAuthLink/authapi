package io

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"testing"
)

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

func TestRequestWith(t *testing.T) {
	t.Run("nil data", func(t *testing.T) {
		var data *any = nil
		req := RequestWith(data)
		if req == nil || req.initialized {
			t.Errorf("expected uninitialized request, got %v", req)
		}
	})

	t.Run("non-nil data", func(t *testing.T) {
		data := "test"
		req := RequestWith(&data)
		if !req.initialized || req.Data != data {
			t.Errorf("expected initialized request with data %s, got %v", data, req)
		}
	})
}

func TestReadRequest(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		var req *Request[any] = nil
		if err := req.Read(nil); err == nil {
			t.Errorf("expected error for nil request, got nil")
		} else if !errors.Is(err, ErrNilRequest) {
			t.Errorf("expected ErrNilRequest, got %v", err)
		}
		if err := req.Read(&http.Request{}); err == nil {
			t.Errorf("expected error for empty request body, got nil")
		} else if !errors.Is(err, ErrNilRequestBody) {
			t.Errorf("expected ErrNilRequestBody, got %v", err)
		}
	})

	t.Run("read body error", func(t *testing.T) {
		errReq := &http.Request{
			Body: io.NopCloser(readerFunc(func([]byte) (int, error) {
				return 0, errors.New("forced read error")
			})),
		}
		var req Request[any]
		if err := req.Read(errReq); err == nil {
			t.Errorf("expected error for read body, got nil")
		} else if !errors.Is(err, ErrReadBody) {
			t.Errorf("expected ErrReadBody, got %v", err)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		errReq := &http.Request{
			Body: io.NopCloser(bytes.NewBuffer([]byte{})),
		}
		var req Request[any]
		if err := req.Read(errReq); err == nil {
			t.Errorf("expected error for empty body, got nil")
		} else if !errors.Is(err, ErrEmptyRequestBody) {
			t.Errorf("expected ErrEmptyRequestBody, got %v", err)
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		errReq := &http.Request{
			Body: io.NopCloser(bytes.NewBuffer([]byte("invalid json"))),
		}
		var req Request[any]
		if err := req.Read(errReq); err == nil {
			t.Errorf("expected error for unmarshal, got nil")
		} else if !errors.Is(err, ErrDecodeBody) {
			t.Errorf("expected ErrDecodeBody, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		type Data struct {
			Message string `json:"message"`
		}
		data := &Data{Message: "Hello, World!"}
		body, _ := json.Marshal(data)
		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		var request Request[Data]
		if err := request.Read(req); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if request.Data.Message != data.Message {
			t.Errorf("expected %s, got %s", data.Message, request.Data.Message)
		}
	})
}

func TestWriteJSONRequest(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		var req *Request[any] = nil
		if err := req.WriteJSON(nil); err == nil {
			t.Errorf("expected error for nil request, got nil")
		} else if !errors.Is(err, ErrNilRequest) {
			t.Errorf("expected ErrNilRequest, got %v", err)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		req := new(Request[any])
		if err := req.WriteJSON(&http.Request{}); err == nil {
			t.Errorf("expected error for not initialized request, got nil")
		} else if !errors.Is(err, ErrNotInitializedRequest) {
			t.Errorf("expected ErrNotInitializedRequest, got %v", err)
		}
	})

	t.Run("write JSON error", func(t *testing.T) {
		data := struct {
			Invalid float64
		}{
			Invalid: math.NaN(),
		}
		req := RequestWith(&data)
		httpReq := &http.Request{
			Header: make(http.Header),
		}
		if err := req.WriteJSON(httpReq); err == nil {
			t.Errorf("expected error for write JSON, got nil")
		} else if !errors.Is(err, ErrEncodeBody) {
			t.Errorf("expected ErrEncodeBody, got %v", err)
		}
		if httpReq.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", httpReq.Header.Get("Content-Type"))
		}
	})

	t.Run("success", func(t *testing.T) {
		data := struct {
			Message string `json:"message"`
		}{
			Message: "Hello, World!",
		}
		req := RequestWith(&data)
		httpReq := &http.Request{
			Header: make(http.Header),
		}
		if err := req.WriteJSON(httpReq); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if httpReq.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", httpReq.Header.Get("Content-Type"))
		}
	})
}

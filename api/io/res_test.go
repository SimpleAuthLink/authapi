package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errorWriter struct {
	*httptest.ResponseRecorder
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("forced write error")
}

func TestResponseWith(t *testing.T) {
	type Data struct {
		Message string `json:"message"`
	}
	nilResp := ResponseWith[Data](nil)
	if !nilResp.empty {
		t.Errorf("expected response to be empty")
	}
	data := &Data{Message: "Hello, World!"}
	resp := ResponseWith(data)
	if resp.empty {
		t.Errorf("expected response to be non-empty")
	}
	if resp.Data.Message != data.Message {
		t.Errorf("expected %s, got %s", data.Message, resp.Data.Message)
	}
}

func TestOkResponse(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		body := []byte("Hello, World!")
		if noEmptyRes := OkResponse(body...); noEmptyRes.empty {
			t.Errorf("expected response to be non-empty")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if emptyRes := OkResponse(); !emptyRes.empty {
			t.Errorf("expected response to be empty")
		}
	})
}

func TestWriteJSON(t *testing.T) {
	type Data struct {
		Message float64 `json:"message"`
	}

	t.Run("encode error", func(t *testing.T) {
		resp := &Response[Data]{Data: Data{Message: math.NaN()}, empty: false}
		rr := httptest.NewRecorder()
		resp.WriteJSON(rr)
		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, status)
		}
	})

	t.Run("write error", func(t *testing.T) {
		resp := &Response[any]{Data: nil, empty: true}
		rr := httptest.NewRecorder()
		resp.WriteJSON(&errorWriter{ResponseRecorder: rr})
		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, status)
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		data := &Data{Message: 1}
		resp := ResponseWith(data)

		rr := httptest.NewRecorder()
		resp.WriteJSON(rr)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, status)
		}

		expected, _ := json.Marshal(data)
		if rr.Body.String() != string(expected)+"\n" { // Ensure newline is accounted for
			t.Errorf("expected body %s, got %s", string(expected)+"\n", rr.Body.String())
		}
	})

	t.Run("empty", func(t *testing.T) {
		// write json with nil data
		nilResp := ResponseWith[string](nil)
		rr := httptest.NewRecorder()
		nilResp.WriteJSON(rr)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, status)
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		resp := OkResponse()
		rr := httptest.NewRecorder()
		resp.Write(rr)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, status)
		}

		if rr.Body.String() != "OK" {
			t.Errorf("expected body OK, got %s", rr.Body.String())
		}
	})

	t.Run("empty error", func(t *testing.T) {
		resp := &Response[any]{empty: true}
		rr := httptest.NewRecorder()
		resp.Write(&errorWriter{ResponseRecorder: rr})

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, status)
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		body := []byte("Hello, World!")
		resp := OkResponse(body...)

		rr := httptest.NewRecorder()
		resp.Write(rr)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, status)
		}

		if rr.Body.String() != string(body) {
			t.Errorf("expected body %s, got %s", string(body), rr.Body.String())
		}
	})

	t.Run("non-empty error", func(t *testing.T) {
		body := []byte("Hello, World!")
		resp := OkResponse(body...)

		rr := httptest.NewRecorder()
		resp.Write(&errorWriter{ResponseRecorder: rr})

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, status)
		}
	})
}

func TestRead(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		if _, err := new(Response[string]).Read(nil); err != ErrNilRequest {
			t.Errorf("expected error %v, got %v", ErrNilRequest, err)
		}
	})

	t.Run("nil body", func(t *testing.T) {
		if _, err := new(Response[string]).Read(&http.Response{Body: nil}); err != ErrNilRequestBody {
			t.Errorf("expected error %v, got %v", ErrNilRequestBody, err)
		}
	})

	t.Run("read error", func(t *testing.T) {
		res := &http.Response{
			Body: io.NopCloser(readerFunc(func([]byte) (int, error) {
				return 0, fmt.Errorf("forced read error")
			})),
		}
		if _, err := new(Response[string]).Read(res); err != ErrReadBody {
			t.Errorf("expected error %v, got %v", ErrReadBody, err)
		}
	})

	t.Run("status not OK", func(t *testing.T) {
		res := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"message":"error"}`))),
		}
		if _, err := new(Response[string]).Read(res); err != ErrHTTPStatusNotOk {
			t.Errorf("expected error %v, got %v", ErrHTTPStatusNotOk, err)
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		res := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte("invalid json"))),
		}
		if _, err := new(Response[string]).Read(res); err != ErrDecodeBody {
			t.Errorf("expected error %v, got %v", ErrDecodeBody, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		type Data struct {
			Message string `json:"message"`
		}
		data := &Data{Message: "Hello, World!"}
		body, _ := json.Marshal(data)
		res := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer(body)),
		}
		result, err := new(Response[Data]).Read(res)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result.Message != data.Message {
			t.Errorf("expected %s, got %s", data.Message, result.Message)
		}
	})
}

func TestBytes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		resp := OkResponse()
		data, ok := resp.bytes()
		if !ok {
			t.Errorf("expected bytes to be valid")
		}
		if data != nil {
			t.Errorf("expected nil data for empty response, got %v", data)
		}
	})

	t.Run("bytes", func(t *testing.T) {
		body := []byte("Hello, World!")
		resp := ResponseWith(&body)
		data, ok := resp.bytes()
		if !ok {
			t.Errorf("expected bytes to be valid")
		}
		if string(data) != string(body) {
			t.Errorf("expected body %s, got %s", string(body), string(data))
		}
	})

	t.Run("string", func(t *testing.T) {
		msg := "Hello, World!"
		resp := ResponseWith(&msg)
		data, ok := resp.bytes()
		if !ok {
			t.Errorf("expected bytes to be valid")
		}
		if string(data) != msg {
			t.Errorf("expected body %s, got %s", msg, string(data))
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		imsg := 123
		resp := ResponseWith(&imsg)
		data, ok := resp.bytes()
		if ok {
			t.Errorf("expected bytes to be invalid")
		}
		if data != nil {
			t.Errorf("expected nil data for invalid response, got %v", data)
		}
	})
}

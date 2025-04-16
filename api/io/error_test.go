package io

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	if err.Code != 1001 {
		t.Errorf("expected code 1001, got %d", err.Code)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
	if err.Message != "" {
		t.Errorf("expected empty message, got %s", err.Message)
	}
	if err.Err != "" {
		t.Errorf("expected empty error string, got %s", err.Err)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	err.Message = "Bad Request"
	err.Err = "Invalid input"
	expected := "code: 1001, message: Bad Request, error: Invalid input, status_code: 400"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestAPIError_WithErr(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	_ = err.WithErr(errors.New("Invalid input"))
	if err.Err != "Invalid input" {
		t.Errorf("expected error string 'Invalid input', got %s", err.Err)
	}

	_ = err.WithErr(errors.New("Missing field"))
	if err.Err != "Invalid input: Missing field" {
		t.Errorf("expected error string 'Invalid input: Missing field', got %s", err.Err)
	}
}

func TestAPIError_With(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	_ = err.With("Bad Request")
	if err.Message != "Bad Request" {
		t.Errorf("expected message 'Bad Request', got %s", err.Message)
	}

	_ = err.With("Invalid input")
	if err.Message != "Bad Request: Invalid input" {
		t.Errorf("expected message 'Bad Request: Invalid input', got %s", err.Message)
	}
}

func TestAPIError_Write(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	err.Message = "Bad Request"
	err.Err = "Invalid input"

	rr := httptest.NewRecorder()
	err.Write(rr)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, status)
	}

	expected := `{"code":1001,"message":"Bad Request","error":"Invalid input"}`
	if rr.Body.String() != expected {
		t.Errorf("expected body %s, got %s", expected, rr.Body.String())
	}
}

func TestAPIError_bytes(t *testing.T) {
	err := NewAPIError(1001, http.StatusBadRequest)
	err.Message = "Bad Request"
	err.Err = "Invalid input"

	data := err.bytes()
	expected := `{"code":1001,"message":"Bad Request","error":"Invalid input"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

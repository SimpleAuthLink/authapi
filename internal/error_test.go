package internal

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewErr(t *testing.T) {
	err := NewErr("test message")
	if err.msg != "test message" {
		t.Errorf("expected message 'test message', got '%s'", err.msg)
	}
	if err.trace != nil {
		t.Errorf("expected nil trace, got '%v'", err.trace)
	}
}

func TestError(t *testing.T) {
	err := NewErr("test message")
	if err.Error() != "test message" {
		t.Errorf("expected 'test message', got '%s'", err.Error())
	}

	wrappedErr := errors.New("wrapped error")
	_ = err.With(wrappedErr)
	expected := "test message: wrapped error"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestWith(t *testing.T) {
	err := NewErr("test message")
	wrappedErr := errors.New("wrapped error")
	_ = err.With(wrappedErr)
	if err.trace != wrappedErr {
		t.Errorf("expected trace '%v', got '%v'", wrappedErr, err.trace)
	}
}

func TestWithf(t *testing.T) {
	err := NewErr("test message")
	_ = err.Withf("formatted %s", "error")
	expectedTrace := fmt.Errorf("formatted %s", "error").Error()
	if err.trace.Error() != expectedTrace {
		t.Errorf("expected trace '%s', got '%s'", expectedTrace, err.trace.Error())
	}
}

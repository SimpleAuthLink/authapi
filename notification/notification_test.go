package notification

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type testConfig struct{}

func (testConfig) Valid() bool {
	return true
}

func (testConfig) Support(n Notification) bool {
	return true
}

type testNotification struct{}

func (testNotification) Valid() bool {
	return true
}

func (testNotification) Send(QueueConfig) error {
	return nil
}

func TestPushQueueFull(t *testing.T) {
	queue, err := NewQueue(context.Background(), 1, testConfig{})
	if err != nil {
		t.Fatalf("init queue: %v", err)
	}
	queue.Start(1)
	defer queue.Stop()

	if err := queue.Push(testNotification{}); err != nil {
		t.Fatalf("first push: %v", err)
	}
	if err := queue.Push(testNotification{}); err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestWorkerPool(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, testConfig{})
	if err != nil {
		t.Fatalf("init queue: %v", err)
	}
	queue.Start(4)
	defer queue.Stop()

	for range 10 {
		if err := queue.Push(testNotification{}); err != nil {
			t.Fatalf("push: %v", err)
		}
	}
}

var errSendFailure = fmt.Errorf("send failure")

type testErrorNotification struct{}

func (testErrorNotification) Valid() bool            { return true }
func (testErrorNotification) Send(QueueConfig) error { return errSendFailure }

func TestErrors(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, testConfig{})
	if err != nil {
		t.Fatalf("init queue: %v", err)
	}

	errCh := queue.Errors()
	if errCh == nil {
		t.Fatal("Errors() returned nil channel")
	}

	queue.Start(1)

	// push a notification that triggers an error
	if err := queue.Push(testErrorNotification{}); err != nil {
		t.Fatalf("push: %v", err)
	}

	// verify the error arrives on the Errors() channel
	select {
	case got := <-errCh:
		if got != errSendFailure {
			t.Errorf("expected %v, got %v", errSendFailure, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected error on Errors() channel")
	}

	// verify Stop() closes the Errors() channel
	queue.Stop()
	_, ok := <-errCh
	if ok {
		t.Error("expected Errors() channel to be closed after Stop()")
	}
}

func TestPushAfterStop(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, testConfig{})
	if err != nil {
		t.Fatal(err)
	}
	queue.Start(1)
	queue.Stop()

	// Push after Stop: returns nil (silent accept)
	if err := queue.Push(testNotification{}); err != nil {
		t.Errorf("expected nil after Stop, got %v", err)
	}
}

type invalidNotification struct{}

func (invalidNotification) Valid() bool            { return false }
func (invalidNotification) Send(QueueConfig) error { return nil }

func TestPushInvalidNotification(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, testConfig{})
	if err != nil {
		t.Fatal(err)
	}
	queue.Start(1)
	defer queue.Stop()

	// Push an invalid notification (always returns false from Valid)
	err = queue.Push(invalidNotification{})
	if err == nil {
		t.Error("expected error for invalid notification, got nil")
	}
}

// Config that Supports nothing — triggers the "no config matches" branch
type noSupportConfig struct{}

func (noSupportConfig) Valid() bool                 { return true }
func (noSupportConfig) Support(n Notification) bool { return false }

// Config that becomes invalid after construction
type invalidOnSendConfig struct{ valid bool }

func (c *invalidOnSendConfig) Valid() bool                 { return c.valid }
func (c *invalidOnSendConfig) Support(n Notification) bool { return true }

func TestSendNoSupport(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, noSupportConfig{})
	if err != nil {
		t.Fatal(err)
	}
	queue.Start(1)
	defer queue.Stop()

	// Sends succeed (no config matches → no error but no delivery)
	if err := queue.Push(testNotification{}); err != nil {
		t.Fatalf("unexpected push error: %v", err)
	}
}

func TestSendInvalidConfig(t *testing.T) {
	cfg := &invalidOnSendConfig{valid: true}
	queue, err := NewQueue(context.Background(), 10, cfg)
	if err != nil {
		t.Fatal(err)
	}
	queue.Start(1)
	defer queue.Stop()

	// Make config invalid mid-flight
	cfg.valid = false

	// Push a notification — worker hits cfg.Valid() → false → error on Errors channel
	if err := queue.Push(testNotification{}); err != nil {
		t.Fatalf("push: %v", err)
	}
	select {
	case got := <-queue.Errors():
		if got == nil {
			t.Error("expected error from send")
		}
	case <-time.After(time.Second):
		t.Error("expected error on Errors() channel")
	}
}

func TestSendSuccess(t *testing.T) {
	queue, err := NewQueue(context.Background(), 10, testConfig{})
	if err != nil {
		t.Fatal(err)
	}
	queue.Start(1)

	if err := queue.Push(testNotification{}); err != nil {
		t.Fatalf("push: %v", err)
	}
	// Worker processes notification via send's happy-path (Send returns nil).
	queue.Stop()
}

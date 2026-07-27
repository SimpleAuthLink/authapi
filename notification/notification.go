package notification

import (
	"context"
	"fmt"
	"sync"
)

// ErrQueueFull is returned when the notification queue is at capacity.
var ErrQueueFull = fmt.Errorf("notification queue full")

// Notification is the interface that must be implemented by any notification
// channel (email, SMS, push, etc.). Implementations must be sendable via a
// QueueConfig that supports them.
type Notification interface {
	// Valid returns true if the notification has all required fields set.
	Valid() bool
	// Send delivers the notification using the given configuration.
	Send(QueueConfig) error
}

// QueueConfig is the interface for notification channel configurations. Each
// channel (email, SMS, etc.) provides its own config that implements this
// interface.
type QueueConfig interface {
	// Valid returns true if the configuration is complete and usable.
	Valid() bool
	// Support reports whether this configuration can send the given
	// notification type.
	Support(Notification) bool
}

// Queue is a generic multi-channel notification queue. It fans out each
// notification to the first QueueConfig that Supports it. Workers are started
// by Start and stopped by Stop. Errors are emitted on the channel returned by
// Errors.
type Queue struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan Notification

	configs  []QueueConfig
	waiter   sync.WaitGroup
	errCh    chan error
	closedMu sync.RWMutex
	closed   bool
}

// NewQueue creates a new Queue with the given buffer size and channel
// configurations. Each config must pass Valid(). Returns an error if any
// config is invalid.
func NewQueue(ctx context.Context, queueSize int, configs ...QueueConfig) (*Queue, error) {
	for _, cfg := range configs {
		if !cfg.Valid() {
			return nil, fmt.Errorf("invalid config provided")
		}
	}

	internalCtx, cancel := context.WithCancel(ctx)
	return &Queue{
		ctx:     internalCtx,
		cancel:  cancel,
		ch:      make(chan Notification, queueSize),
		configs: configs,
		errCh:   make(chan error, 1),
	}, nil
}

// Start launches workers goroutines that consume the queue. Each worker reads
// notifications and sends them via the matching config. Send errors are pushed
// to the Errors channel (non-blocking).
func (eq *Queue) Start(workers int) {
	for range workers {
		eq.waiter.Go(func() {
			for {
				select {
				case <-eq.ctx.Done():
					return
				case e, ok := <-eq.ch:
					if !ok {
						return
					}
					if err := eq.send(e); err != nil {
						select {
						case eq.errCh <- err:
						default:
						}
					}
				}
			}
		})
	}
}

// Stop gracefully shuts down the queue: closes the input channel, cancels the
// context, waits for all workers to finish, and closes the Errors channel. New
// pushes after Stop are silently accepted (return nil).
func (eq *Queue) Stop() {
	eq.closedMu.Lock()
	eq.closed = true
	close(eq.ch)
	eq.closedMu.Unlock()
	eq.cancel()
	eq.waiter.Wait()
	close(eq.errCh)
}

// Push enqueues a notification. Returns ErrQueueFull if the buffer is full, or
// nil if the notification was accepted (including after Stop).
func (eq *Queue) Push(n Notification) error {
	if !n.Valid() {
		return fmt.Errorf("invalid notification")
	}
	eq.closedMu.RLock()
	defer eq.closedMu.RUnlock()
	if eq.closed {
		return nil
	}
	select {
	case eq.ch <- n:
		return nil
	default:
		return ErrQueueFull
	}
}

// Errors returns a receive-only channel that emits errors encountered
// by workers during notification sends. The caller must continuously
// read from this channel after calling Start to prevent workers from
// blocking. The channel is closed when Stop is called.
func (eq *Queue) Errors() <-chan error {
	return eq.errCh
}

func (eq *Queue) send(n Notification) error {
	for _, cfg := range eq.configs {
		if !cfg.Support(n) {
			continue
		}
		if !cfg.Valid() {
			return fmt.Errorf("invalid configuration")
		}
		if err := n.Send(cfg); err != nil {
			return err
		}
	}
	return nil
}

package cosyne

import (
	"context"
	"sync"
)

// Signal lets one goroutine notify another that something has happened.
//
// A signal can be notified repeatedly. Each notification unblocks at most one
// waiting goroutine and is consumed by it. A notification that arrives while
// another is already pending is coalesced into it.
type Signal struct {
	once sync.Once
	ch   chan struct{}
}

// SignalR is the read-only view of a [Signal]. It cannot notify the signal.
//
// If several goroutines wait on the same SignalR, each notification unblocks
// only one of them.
type SignalR interface {
	Wait()
	WaitContext(ctx context.Context) error
	Chan() <-chan struct{}

	sealed()
}

// SignalW is the write-only view of a [Signal]. It cannot wait on the signal.
type SignalW interface {
	Notify() bool

	sealed()
}

var (
	_ SignalR = (*Signal)(nil)
	_ SignalW = (*Signal)(nil)
)

// Notify unblocks one waiting goroutine.
//
// If no goroutine is waiting, the notification remains pending and the next
// goroutine to wait unblocks immediately.
//
// It returns true if the notification was recorded, or false if one was already
// pending.
func (s *Signal) Notify() bool {
	select {
	case s.init() <- struct{}{}:
		return true
	default:
		return false
	}
}

// Wait blocks until the signal is notified, consuming the notification.
func (s *Signal) Wait() {
	<-s.init()
}

// WaitContext blocks until the signal is notified or ctx is done.
//
// It returns nil if it consumed a notification, or ctx's error otherwise. If
// ctx is already done, it returns its error without consuming a pending
// notification.
func (s *Signal) WaitContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.init():
		return nil
	}
}

// Chan returns a channel that unblocks when the signal is notified.
//
// Receiving from the channel consumes the notification, as [Signal.Wait] does.
func (s *Signal) Chan() <-chan struct{} {
	return s.init()
}

func (*Signal) sealed() {}

func (s *Signal) init() chan struct{} {
	s.once.Do(func() {
		s.ch = make(chan struct{}, 1)
	})

	return s.ch
}

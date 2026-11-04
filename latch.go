package cosyne

import (
	"context"
	"sync"
)

// Latch is a one-way flag that unblocks any goroutines waiting on it when it is
// set. A latch cannot be reset.
type Latch struct {
	initOnce, setOnce sync.Once
	ch                chan struct{}
}

// LatchR is the read-only view of a [Latch]. It cannot set the latch.
type LatchR interface {
	IsSet() bool
	Wait()
	WaitContext(ctx context.Context) error
	Chan() <-chan struct{}

	sealed()
}

// LatchW is the write-only view of a [Latch]. It cannot observe the latch.
type LatchW interface {
	Set() bool

	sealed()
}

var (
	_ LatchR = (*Latch)(nil)
	_ LatchW = (*Latch)(nil)
)

// Set unblocks any goroutines that are waiting on the latch.
//
// It returns true if the latch was set by this call, or false if it was already
// set.
func (l *Latch) Set() bool {
	set := false

	l.setOnce.Do(func() {
		close(l.init())
		set = true
	})

	return set
}

// IsSet reports whether the latch has been set.
func (l *Latch) IsSet() bool {
	select {
	case <-l.init():
		return true
	default:
		return false
	}
}

// Wait blocks until the latch is set.
func (l *Latch) Wait() {
	<-l.init()
}

// WaitContext blocks until the latch is set or ctx is done.
//
// It returns nil if the latch is set, or ctx's error otherwise.
func (l *Latch) WaitContext(ctx context.Context) error {
	if l.IsSet() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.init():
		return nil
	}
}

// Chan returns a channel that is closed when the latch is set.
func (l *Latch) Chan() <-chan struct{} {
	return l.init()
}

func (l *Latch) init() chan struct{} {
	l.initOnce.Do(func() {
		l.ch = make(chan struct{})
	})

	return l.ch
}

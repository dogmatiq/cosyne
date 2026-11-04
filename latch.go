package cosyne

import (
	"context"
	"sync"
	"sync/atomic"
)

// Latch is a one-way flag that unblocks any goroutines waiting on it when it is
// set. A latch cannot be reset.
type Latch struct {
	hasChan, isSet atomic.Bool
	m              sync.Mutex
	ch             chan struct{}
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
	if l.isSet.Load() {
		return false
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.isSet.Load() {
		return false
	}

	if l.ch == nil {
		l.ch = make(chan struct{})
		l.hasChan.Store(true)
	}

	close(l.ch)
	l.isSet.Store(true)

	return true
}

// IsSet reports whether the latch has been set.
func (l *Latch) IsSet() bool {
	return l.isSet.Load()
}

// Wait blocks until the latch is set.
func (l *Latch) Wait() {
	<-l.Chan()
}

// WaitContext blocks until the latch is set or ctx is done.
//
// It returns nil if the latch is set, or ctx's error otherwise.
func (l *Latch) WaitContext(ctx context.Context) error {
	if l.isSet.Load() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.Chan():
		return nil
	}
}

// Chan returns a channel that is closed when the latch is set.
func (l *Latch) Chan() <-chan struct{} {
	if l.hasChan.Load() {
		return l.ch
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.ch == nil {
		l.ch = make(chan struct{})
		l.hasChan.Store(true)
	}

	return l.ch
}

func (*Latch) sealed() {}

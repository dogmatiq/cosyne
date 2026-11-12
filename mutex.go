package cosyne

import (
	"context"
	"sync"
)

// Mutex is a mutual exclusion lock that can be acquired with a context.
//
// Unlike [sync.Mutex], a call to [Mutex.Lock] stops waiting when its context is
// done.
type Mutex struct {
	// once creates ch on first use.
	once sync.Once

	// ch holds a value while the mutex is locked.
	ch chan struct{}
}

// Lock acquires the mutex, blocking until it's available or ctx is done.
//
// It returns nil if it acquired the mutex, or ctx's error otherwise. If ctx is
// already done, it returns its error without acquiring the mutex.
func (m *Mutex) Lock(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.init() <- struct{}{}:
		return nil
	}
}

// TryLock acquires the mutex if it's available, without blocking.
//
// It returns true if it acquired the mutex, or false if the mutex is already
// locked.
func (m *Mutex) TryLock() bool {
	select {
	case m.init() <- struct{}{}:
		return true
	default:
		return false
	}
}

// Unlock releases the mutex.
//
// It panics if the mutex isn't locked.
func (m *Mutex) Unlock() {
	select {
	case <-m.init():
	default:
		panic("cosyne: unlock of unlocked mutex")
	}
}

func (m *Mutex) init() chan struct{} {
	m.once.Do(func() {
		m.ch = make(chan struct{}, 1)
	})

	return m.ch
}

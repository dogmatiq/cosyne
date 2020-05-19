package cosyne

import (
	"context"
	"sync"
)

// Mutex is a context-aware mutex.
type Mutex struct {
	once  sync.Once
	guard chan struct{} // buffered guard, write = lock, read = unlock
}

// Lock acquires an exclusive lock on the mutex.
//
// It blocks until the mutex is acquired, or ctx is canceled.
func (m *Mutex) Lock(ctx context.Context) error {
	m.once.Do(func() {
		m.guard = make(chan struct{}, 1)
	})

	select {
	case m.guard <- struct{}{}: // lock the mutex
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryLock acquires an exclusive lock on the mutex if doing so would not block.
//
// It returns true if the mutex is locked.
func (m *Mutex) TryLock() bool {
	m.once.Do(func() {
		m.guard = make(chan struct{}, 1)
	})

	select {
	case m.guard <- struct{}{}: // lock the mutex
		return true
	default:
		return false
	}
}

// Unlock releases the mutex.
//
// It panics if the mutex is not currently locked.
func (m *Mutex) Unlock() {
	m.once.Do(func() {
		m.guard = make(chan struct{}, 1)
	})

	select {
	case <-m.guard: // unlock the mutex
		return // keep to see coverage
	default:
		panic("mutex is not locked")
	}
}

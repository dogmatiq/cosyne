package cosyne

import (
	"context"
	"sync"
)

// RWMutex is a context-aware read/write mutex.
type RWMutex struct {
	m         sync.Mutex
	readers   int // negative = write lock acquired
	unlockedC chan struct{}
	retryC    chan struct{}
}

// Lock acquires an exclusive lock on the mutex.
//
// It blocks until the mutex is acquired, or ctx is canceled.
func (m *RWMutex) Lock(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	m.m.Lock()

	unlocked := m.unlocked()
	if unlocked == nil {
		m.readers--
		m.m.Unlock()
		return nil
	}

	m.m.Unlock()

	// Since we need an exclusive lock, we don't care how many readers/writers
	// there are, we just want to know when it's our turn.

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-unlocked:
		// We've obtained exclusive access, mark the mutex as "write-locked" by
		// sending the reader count negative.
		m.m.Lock()
		m.readers--
		m.m.Unlock()

		return nil
	}
}

// TryLock acquires an exclusive lock on the mutex if doing so would not block.
//
// It returns true if the mutex is locked.
func (m *RWMutex) TryLock() bool {
	m.m.Lock()
	defer m.m.Unlock()

	if m.unlocked() == nil {
		m.readers--
		return true
	}

	return false
}

// Unlock releases the mutex.
//
// It panics if the mutex is not currently locked with Lock().
func (m *RWMutex) Unlock() {
	m.m.Lock()

	if m.readers >= 0 {
		m.m.Unlock()
		panic("mutex is not write-locked")
	}

	m.readers++
	m.signalUnlocked()

	m.m.Unlock()
}

// RLock acquires a shared lock on the mutex.
//
// It blocks until the mutex is acquired, or ctx is canceled.
func (m *RWMutex) RLock(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	for {
		m.m.Lock()

		// If there are already other readers, just add ourselves to the reader
		// count immediately.
		if m.readers > 0 {
			m.readers++
			m.m.Unlock()
			return nil
		}

		// Otherwise, we need to wait until we have exclusive access in order to
		// "convert" the mutex to read-locked.
		//
		// If we get it straight away we tell any other blocking RLock() calls
		// to retry.
		unlocked := m.unlocked()
		if unlocked == nil {
			m.readers++
			m.signalRetry()
			m.m.Unlock()
			return nil
		}

		// We also need to be notified when to retry if some other read-locker
		// gets exclusive access before us.
		retry := m.retry()

		// Release the internal mutex before waiting for exclusive access.
		m.m.Unlock()

		// And now we wait ...
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-retry:
			// Some other blocking call to RLock() obtained exclusive access
			// first, and notified us that the mutex is ready for reads.
			//
			// We have to retry from the beginning, in case the read-lock has
			// already been released.
			continue

		case <-unlocked:
			// We've obtained exclusive access, mark the mutex as "read-locked"
			// by sending the reader count positive.
			//
			// We then tell any other blocking RLock() calls to retry.
			m.m.Lock()
			m.readers++
			m.signalRetry()
			m.m.Unlock()

			return nil
		}
	}
}

// TryRLock acquires a shared lock on the mutex if doing so would not block.
//
// It returns true if the mutex is locked.
func (m *RWMutex) TryRLock() bool {
	m.m.Lock()
	defer m.m.Unlock()

	// If there are already other readers, just add ourselves to the reader
	// count immediately.
	if m.readers > 0 {
		m.readers++
		return true
	}

	// Otherwise, we need to check if we we have exclusive access in order to
	// "convert" the mutex to read-locked.
	if m.unlocked() == nil {
		m.readers++
		m.signalRetry()
		return true
	}

	return false
}

// RUnlock releases the mutex.
//
// It panics if the mutex is not currently locked with RLock().
func (m *RWMutex) RUnlock() {
	m.m.Lock()

	if m.readers <= 0 {
		m.m.Unlock()
		panic("mutex is not read-locked")
	}

	m.readers--

	if m.readers == 0 {
		m.signalUnlocked()
	}

	m.m.Unlock()
}

// unlocked returns a channel used to signal to a single consumer that the mutex
// has been unlocked.
//
// It assumes m.m is locked.
//
// It returns nil if the m itself is already unlocked.
func (m *RWMutex) unlocked() <-chan struct{} {
	if m.readers != 0 {
		// The mutex is locked, we need to wait regardless of whether it's
		// Lock()'d or RLock()'d.
		return m.unlockedC
	}

	if m.unlockedC == nil {
		// This is the first time the mutex has been locked. Create the buffered
		// channel but don't write anything to it.
		m.unlockedC = make(chan struct{}, 1)
	} else {
		// Otherwise, the channel already exists but the mutex is unlocked so
		// reading will not block.
		<-m.unlockedC
	}

	return nil
}

// signalUnlocked signals that the mutex has been unlocked, waking a single
// waiting goroutine, if present.
//
// It assumes m.m is locked.
func (m *RWMutex) signalUnlocked() {
	m.unlockedC <- struct{}{}
}

// retry returns a channel that is closed when a goroutine that is waiting to
// obtain a read-lock should retry.
//
// It assumes m.m is locked.
func (m *RWMutex) retry() <-chan struct{} {
	if m.retryC == nil {
		m.retryC = make(chan struct{})
	}

	return m.retryC
}

// signalRetry wakes any goroutines that are awaiting to retry obtaining a
// read-lock.
//
// It assumes m.m is locked.
func (m *RWMutex) signalRetry() {
	// If m.retry is already nil, it means that a competing goroutine has
	// already closed it AND called RUnlock() and we happened to see the send to
	// m.unlockedC before the closure of m.retryC.
	//
	// See https://github.com/dogmatiq/infix/issues/72.
	if m.retryC != nil {
		close(m.retryC)
		m.retryC = nil
	}
}

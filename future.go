package cosyne

import (
	"context"
	"sync"
	"sync/atomic"
)

// Future holds a value that is stored at most once, and may not be available
// yet.
type Future[T any] struct {
	ptr  atomic.Pointer[T]
	once sync.Once
	ch   chan struct{}
}

// FutureR is the read-only view of a [Future]. It cannot store a value.
type FutureR[T any] interface {
	Load() (T, bool)
	Wait() T
	WaitContext(ctx context.Context) (T, error)
	Chan() <-chan struct{}

	sealed()
}

// FutureW is the write-only view of a [Future]. It cannot observe the value.
type FutureW[T any] interface {
	Store(v T) bool

	sealed()
}

var (
	_ FutureR[int] = (*Future[int])(nil)
	_ FutureW[int] = (*Future[int])(nil)
)

// Store sets the value of the future.
//
// It returns true if the value was set by this call, or false if it was already
// set, in which case the existing value is unchanged.
func (f *Future[T]) Store(v T) bool {
	if f.ptr.CompareAndSwap(nil, &v) {
		close(f.init())
		return true
	}

	return false
}

// Load returns the value of the future, and true, if it is available.
//
// Otherwise, it returns the zero value and false.
func (f *Future[T]) Load() (T, bool) {
	if p := f.ptr.Load(); p != nil {
		return *p, true
	}

	var zero T
	return zero, false
}

// Wait blocks until the value is available, then returns it.
func (f *Future[T]) Wait() T {
	<-f.init()
	return *f.ptr.Load()
}

// WaitContext blocks until the value is available or ctx is done.
//
// It returns the value if it is available, or ctx's error otherwise.
func (f *Future[T]) WaitContext(ctx context.Context) (T, error) {
	if p := f.ptr.Load(); p != nil {
		return *p, nil
	}

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.init():
		return *f.ptr.Load(), nil
	}
}

// Chan returns a channel that is closed when the value becomes available.
func (f *Future[T]) Chan() <-chan struct{} {
	return f.init()
}

func (f *Future[T]) init() chan struct{} {
	f.once.Do(func() {
		f.ch = make(chan struct{})
	})

	return f.ch
}

package cosyne

import "sync/atomic"

// Observable is an atomic value that can be observed for changes.
//
// Its initial value is the zero value of T.
type Observable[T any] struct {
	state atomic.Pointer[observableState[T]]
}

// ObservableR is the read-only view of an [Observable]. It cannot store a
// value.
type ObservableR[T any] interface {
	Load() (T, <-chan struct{})

	sealed()
}

// ObservableW is the write-only view of an [Observable]. It cannot observe the
// value.
type ObservableW[T any] interface {
	Store(v T)

	sealed()
}

var (
	_ ObservableR[int] = (*Observable[int])(nil)
	_ ObservableW[int] = (*Observable[int])(nil)
)

// Load returns the current value, and a channel that is closed when that value
// is replaced by a subsequent call to [Observable.Store].
func (o *Observable[T]) Load() (T, <-chan struct{}) {
	if s := o.state.Load(); s != nil {
		return s.value, s.replaced
	}

	s := &observableState[T]{
		replaced: make(chan struct{}),
	}

	if o.state.CompareAndSwap(nil, s) {
		return s.value, s.replaced
	}

	s = o.state.Load()
	return s.value, s.replaced
}

// Store replaces the current value.
func (o *Observable[T]) Store(v T) {
	prev := o.state.Swap(&observableState[T]{
		value:    v,
		replaced: make(chan struct{}),
	})

	if prev != nil {
		close(prev.replaced)
	}
}

func (*Observable[T]) sealed() {}

type observableState[T any] struct {
	value    T
	replaced chan struct{}
}

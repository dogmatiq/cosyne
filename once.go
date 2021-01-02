package cosyne

import (
	"context"
	"sync/atomic"
)

// Once is a context-aware and "failable" version of sync.Once.
type Once struct {
	done uint32 // atomic bool
	m    Mutex
}

// Do calls the function fn if and only if Do() has never been called
// successfully for this instance of Once.
//
// A successful call is one that returns a nil error and does not panic.
func (o *Once) Do(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if atomic.LoadUint32(&o.done) == 0 {
		return o.doSlow(ctx, fn)
	}

	return nil
}

func (o *Once) doSlow(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := o.m.Lock(ctx); err != nil {
		return err
	}
	defer o.m.Unlock()

	if o.done == 0 {
		if err := fn(ctx); err != nil {
			return err
		}

		atomic.StoreUint32(&o.done, 1)
	}

	return nil
}

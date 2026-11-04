package cosyne

import (
	"context"
	"sync"
	"sync/atomic"
)

// Once is a context-aware variant of [sync.Once] that allows the operation to
// fail. The operation is repeated until it succeeds.
type Once struct {
	done atomic.Bool
	init sync.Once
	busy chan struct{}
}

// Do calls fn if and only if it has not already been called successfully.
//
// A call is successful if fn returns a nil error and does not panic. If it
// fails, a later call to Do calls fn again.
//
// fn is called with ctx, and at most one call to fn is in progress at a time.
// Do blocks while another call is in progress, until that call completes or ctx
// is done.
//
// It returns nil if fn has succeeded, otherwise the error returned by fn, or
// ctx's error.
func (o *Once) Do(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if o.done.Load() {
		return nil
	}

	busy := o.guard()

	select {
	case busy <- struct{}{}:
		defer func() { <-busy }()
	case <-ctx.Done():
		return ctx.Err()
	}

	if o.done.Load() {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		return err
	}

	o.done.Store(true)

	return nil
}

func (o *Once) guard() chan struct{} {
	o.init.Do(func() {
		o.busy = make(chan struct{}, 1)
	})

	return o.busy
}

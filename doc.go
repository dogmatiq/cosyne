// Package cosyne provides context-aware synchronization primitives.
//
// Every primitive is ready to use as its zero value and must not be copied
// after first use. [Signal], [Latch] and [Future] can be waited for with or
// without a [context.Context], or used in a select statement via their Chan
// method.
//
// If a method waits with a context and the awaited state has already been
// reached, it succeeds even if the context is done. Otherwise, a context that
// is already done fails the call before anything is consumed or run. [Signal]
// is the exception: receiving consumes a notification, so a done context always
// wins and any pending notification is left in place.
//
// Where a method reports success as a bool, true means that the call took
// effect and false means that the state was already as requested.
//
// Interfaces whose names end in R or W, such as [LatchR] and [LatchW], are the
// read-only and write-only views of a type. They can only be implemented by the
// types in this package.
package cosyne

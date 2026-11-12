package cosyne_test

import (
	"context"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestMutex(t *testing.T) {
	t.Run("func Lock()", func(t *testing.T) {
		t.Run("when the mutex is unlocked", func(t *testing.T) {
			t.Run("it acquires the mutex", func(t *testing.T) {
				var mutex Mutex

				if err := mutex.Lock(t.Context()); err != nil {
					t.Fatal(err)
				}

				if mutex.TryLock() {
					t.Fatal("expected the mutex to be locked")
				}
			})

		})

		t.Run("when the mutex is unlocked and the context is already done", func(t *testing.T) {
			t.Run("it returns the context error without acquiring the mutex", func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				// Repeat to catch a nondeterministic choice between the two.
				for range 100 {
					var mutex Mutex

					if err := mutex.Lock(ctx); err != context.Canceled {
						t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
					}

					if !mutex.TryLock() {
						t.Fatal("expected the mutex to be unlocked")
					}
				}
			})
		})

		t.Run("when the mutex is unlocked while waiting", func(t *testing.T) {
			t.Run("it acquires the mutex", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var mutex Mutex
					mutex.TryLock()

					returned := false
					var err error
					go func() {
						err = mutex.Lock(t.Context())
						returned = true
					}()

					synctest.Wait()

					if returned {
						t.Fatal("expected Lock() to block while the mutex is locked")
					}

					mutex.Unlock()
					synctest.Wait()

					if !returned {
						t.Fatal("expected Lock() to return")
					}

					if err != nil {
						t.Fatal(err)
					}

					if mutex.TryLock() {
						t.Fatal("expected the mutex to be locked")
					}
				})
			})
		})

		t.Run("when the context is canceled while waiting", func(t *testing.T) {
			t.Run("it returns the context error", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var mutex Mutex
					mutex.TryLock()

					ctx, cancel := context.WithCancel(t.Context())

					var err error
					go func() {
						err = mutex.Lock(ctx)
					}()

					synctest.Wait()
					cancel()
					synctest.Wait()

					if err != context.Canceled {
						t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
					}
				})
			})
		})
	})

	t.Run("func TryLock()", func(t *testing.T) {
		t.Run("when the mutex is unlocked", func(t *testing.T) {
			t.Run("it acquires the mutex and returns true", func(t *testing.T) {
				var mutex Mutex

				if !mutex.TryLock() {
					t.Fatal("expected TryLock() to acquire the mutex")
				}

				if mutex.TryLock() {
					t.Fatal("expected the mutex to be locked")
				}
			})
		})

		t.Run("when the mutex is locked", func(t *testing.T) {
			t.Run("it returns false", func(t *testing.T) {
				var mutex Mutex
				mutex.TryLock()

				if mutex.TryLock() {
					t.Fatal("expected TryLock() to fail")
				}
			})
		})
	})

	t.Run("func Unlock()", func(t *testing.T) {
		t.Run("when the mutex is locked", func(t *testing.T) {
			t.Run("it allows the mutex to be acquired again", func(t *testing.T) {
				var mutex Mutex
				mutex.TryLock()
				mutex.Unlock()

				if !mutex.TryLock() {
					t.Fatal("expected the mutex to be unlocked")
				}
			})
		})

		t.Run("when the mutex is unlocked", func(t *testing.T) {
			t.Run("it panics", func(t *testing.T) {
				var mutex Mutex

				defer func() {
					if recover() == nil {
						t.Fatal("expected a panic")
					}
				}()

				mutex.Unlock()
			})
		})
	})
}

package cosyne_test

import (
	"context"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestLatch(t *testing.T) {
	t.Run("func Set()", func(t *testing.T) {
		t.Run("it unblocks existing waiters", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var latch Latch

				unblockedWait := false
				go func() {
					latch.Wait()
					unblockedWait = true
				}()

				unblockedChan := false
				go func() {
					<-latch.Chan()
					unblockedChan = true
				}()

				// Block until the waiting goroutines are blocked on the latch.
				synctest.Wait()

				latch.Set()

				// Block until the waiting goroutines exit (or remain blocked if
				// something is wrong).
				synctest.Wait()

				if !unblockedWait {
					t.Fatal("expected Wait() to unblock")
				}

				if !unblockedChan {
					t.Fatal("expected Chan() to unblock")
				}
			})
		})

		t.Run("it unblocks future waiters", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var latch Latch

				latch.Set()

				unblockedWait := false
				go func() {
					latch.Wait()
					unblockedWait = true
				}()

				unblockedChan := false
				go func() {
					<-latch.Chan()
					unblockedChan = true
				}()

				synctest.Wait()

				if !unblockedWait {
					t.Fatal("expected Wait() to unblock")
				}

				if !unblockedChan {
					t.Fatal("expected Chan() to unblock")
				}
			})
		})

		t.Run("it returns true only when the latch was not already set", func(t *testing.T) {
			var latch Latch

			if !latch.Set() {
				t.Fatal("expected first call to set the latch")
			}

			if latch.Set() {
				t.Fatal("expected second call to find the latch already set")
			}
		})
	})

	t.Run("func IsSet()", func(t *testing.T) {
		t.Run("it reports whether the latch has been set", func(t *testing.T) {
			var latch Latch

			if latch.IsSet() {
				t.Fatal("did not expect latch to be set")
			}

			latch.Set()

			if !latch.IsSet() {
				t.Fatal("expected latch to be set")
			}
		})
	})

	t.Run("func WaitContext()", func(t *testing.T) {
		t.Run("when the latch is set while waiting", func(t *testing.T) {
			t.Run("it returns nil", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var latch Latch

					var err error
					go func() {
						err = latch.WaitContext(context.Background())
					}()

					synctest.Wait()
					latch.Set()
					synctest.Wait()

					if err != nil {
						t.Fatal(err)
					}
				})
			})
		})

		t.Run("when the context is canceled while waiting", func(t *testing.T) {
			t.Run("it returns the context error", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var latch Latch

					ctx, cancel := context.WithCancel(t.Context())

					var err error
					go func() {
						err = latch.WaitContext(ctx)
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

		t.Run("when the latch is already set and the context is done", func(t *testing.T) {
			t.Run("it returns nil", func(t *testing.T) {
				var latch Latch
				latch.Set()

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				// Repeat to catch a nondeterministic choice between the two.
				for range 100 {
					if err := latch.WaitContext(ctx); err != nil {
						t.Fatal(err)
					}
				}
			})
		})

		t.Run("when the latch is not set and the context is done", func(t *testing.T) {
			t.Run("it returns the context error", func(t *testing.T) {
				var latch Latch

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				if err := latch.WaitContext(ctx); err != context.Canceled {
					t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
				}
			})
		})
	})

	t.Run("func LatchW", func(t *testing.T) {
		t.Run("it sets the latch it views", func(t *testing.T) {
			var latch Latch
			var w LatchW = &latch

			w.Set()

			if !latch.IsSet() {
				t.Fatal("expected latch to be set")
			}
		})
	})
}

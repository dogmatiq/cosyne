package cosyne_test

import (
	"context"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestSignal(t *testing.T) {
	t.Run("func Notify()", func(t *testing.T) {
		t.Run("it unblocks an existing waiter", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var signal Signal

				unblocked := false
				go func() {
					signal.Wait()
					unblocked = true
				}()

				// Block until the waiting goroutine is blocked on the signal.
				synctest.Wait()

				if !signal.Notify() {
					t.Fatal("expected notification to be recorded")
				}

				// Block until the waiting goroutine exits (or remains blocked
				// if something is wrong).
				synctest.Wait()

				if !unblocked {
					t.Fatal("expected Wait() to unblock")
				}
			})
		})

		t.Run("it unblocks a future waiter", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var signal Signal

				signal.Notify()

				unblockedWait := false
				go func() {
					signal.Wait()
					unblockedWait = true
				}()

				synctest.Wait()

				if !unblockedWait {
					t.Fatal("expected Wait() to unblock")
				}

				signal.Notify()

				unblockedChan := false
				go func() {
					<-signal.Chan()
					unblockedChan = true
				}()

				synctest.Wait()

				if !unblockedChan {
					t.Fatal("expected Chan() to unblock")
				}
			})
		})

		t.Run("it unblocks only one waiter per notification", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var signal Signal

				unblocked := 0
				for range 2 {
					go func() {
						signal.Wait()
						unblocked++
					}()
				}

				synctest.Wait()
				signal.Notify()
				synctest.Wait()

				if unblocked != 1 {
					t.Fatalf("unexpected number of unblocked waiters: got %d, want 1", unblocked)
				}

				signal.Notify()
				synctest.Wait()

				if unblocked != 2 {
					t.Fatalf("unexpected number of unblocked waiters: got %d, want 2", unblocked)
				}
			})
		})

		t.Run("it coalesces notifications while one is pending", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var signal Signal

				if !signal.Notify() {
					t.Fatal("expected first notification to be recorded")
				}

				if signal.Notify() {
					t.Fatal("expected second notification to be coalesced")
				}

				signal.Wait()

				unblocked := false
				go func() {
					signal.Wait()
					unblocked = true
				}()

				synctest.Wait()

				if unblocked {
					t.Fatal("expected only one notification to have been pending")
				}

				signal.Notify()
				synctest.Wait()

				if !unblocked {
					t.Fatal("expected Wait() to unblock")
				}
			})
		})
	})

	t.Run("func WaitContext()", func(t *testing.T) {
		t.Run("when the signal is notified while waiting", func(t *testing.T) {
			t.Run("it returns nil", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var signal Signal

					var err error
					go func() {
						err = signal.WaitContext(context.Background())
					}()

					synctest.Wait()
					signal.Notify()
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
					var signal Signal

					ctx, cancel := context.WithCancel(t.Context())

					var err error
					go func() {
						err = signal.WaitContext(ctx)
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

		t.Run("when a notification is pending and the context is done", func(t *testing.T) {
			t.Run("it returns the context error and leaves the notification pending", func(t *testing.T) {
				var signal Signal

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				// Repeat to catch a nondeterministic choice between the two.
				for range 100 {
					signal.Notify()

					if err := signal.WaitContext(ctx); err != context.Canceled {
						t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
					}

					// A pending notification coalesces any further one, so
					// Notify() reports whether the earlier one was consumed.
					if signal.Notify() {
						t.Fatal("expected the notification to remain pending")
					}

					signal.Wait()
				}
			})
		})
	})

	t.Run("func SignalW", func(t *testing.T) {
		t.Run("it notifies the signal it views", func(t *testing.T) {
			var signal Signal
			var w SignalW = &signal

			if !w.Notify() {
				t.Fatal("expected notification to be recorded")
			}

			// The notification is already pending, so waiting does not block.
			signal.Wait()
		})
	})
}

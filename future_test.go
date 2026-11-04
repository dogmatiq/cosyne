package cosyne_test

import (
	"context"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestFuture(t *testing.T) {
	t.Run("func Store()", func(t *testing.T) {
		t.Run("it keeps the first value", func(t *testing.T) {
			var f Future[int]

			if !f.Store(1) {
				t.Fatal("expected first call to store the value")
			}

			if f.Store(2) {
				t.Fatal("expected second call to find a value already stored")
			}

			if v, _ := f.Load(); v != 1 {
				t.Fatalf("unexpected value: got %d, want 1", v)
			}
		})
	})

	t.Run("func Load()", func(t *testing.T) {
		t.Run("it reports whether the value is available", func(t *testing.T) {
			var f Future[int]

			if v, ok := f.Load(); ok || v != 0 {
				t.Fatalf("unexpected result before store: got (%d, %t), want (0, false)", v, ok)
			}

			f.Store(42)

			if v, ok := f.Load(); !ok || v != 42 {
				t.Fatalf("unexpected result after store: got (%d, %t), want (42, true)", v, ok)
			}
		})
	})

	t.Run("func Wait()", func(t *testing.T) {
		t.Run("it returns the value once it is stored", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var f Future[int]

				var got int
				go func() {
					got = f.Wait()
				}()

				synctest.Wait()
				f.Store(42)
				synctest.Wait()

				if got != 42 {
					t.Fatalf("unexpected value: got %d, want 42", got)
				}
			})
		})
	})

	t.Run("func Chan()", func(t *testing.T) {
		t.Run("it is closed once the value is stored", func(t *testing.T) {
			var f Future[int]

			select {
			case <-f.Chan():
				t.Fatal("did not expect channel to be closed")
			default:
			}

			f.Store(42)

			select {
			case <-f.Chan():
			default:
				t.Fatal("expected channel to be closed")
			}
		})
	})

	t.Run("func WaitContext()", func(t *testing.T) {
		t.Run("when the value is stored while waiting", func(t *testing.T) {
			t.Run("it returns the value", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var f Future[int]

					var (
						got int
						err error
					)
					go func() {
						got, err = f.WaitContext(context.Background())
					}()

					synctest.Wait()
					f.Store(42)
					synctest.Wait()

					if err != nil {
						t.Fatal(err)
					}

					if got != 42 {
						t.Fatalf("unexpected value: got %d, want 42", got)
					}
				})
			})
		})

		t.Run("when the context is canceled while waiting", func(t *testing.T) {
			t.Run("it returns the context error", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var f Future[int]

					ctx, cancel := context.WithCancel(t.Context())

					var err error
					go func() {
						_, err = f.WaitContext(ctx)
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

		t.Run("when the value is already stored and the context is done", func(t *testing.T) {
			t.Run("it returns the value", func(t *testing.T) {
				var f Future[int]
				f.Store(42)

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				got, err := f.WaitContext(ctx)
				if err != nil {
					t.Fatal(err)
				}

				if got != 42 {
					t.Fatalf("unexpected value: got %d, want 42", got)
				}
			})
		})

		t.Run("when the value is not stored and the context is done", func(t *testing.T) {
			t.Run("it returns the context error", func(t *testing.T) {
				var f Future[int]

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				if _, err := f.WaitContext(ctx); err != context.Canceled {
					t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
				}
			})
		})
	})

	t.Run("func FutureW", func(t *testing.T) {
		t.Run("it stores the value in the future it views", func(t *testing.T) {
			var f Future[int]
			var w FutureW[int] = &f

			w.Store(42)

			if v, ok := f.Load(); !ok || v != 42 {
				t.Fatalf("unexpected result: got (%d, %t), want (42, true)", v, ok)
			}
		})
	})
}

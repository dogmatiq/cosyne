package cosyne_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestOnce(t *testing.T) {
	t.Run("func Do()", func(t *testing.T) {
		t.Run("it does not call the function again after success", func(t *testing.T) {
			var once Once

			if err := once.Do(
				t.Context(),
				func(context.Context) error {
					return nil
				},
			); err != nil {
				t.Fatal(err)
			}

			if err := once.Do(
				t.Context(),
				func(context.Context) error {
					t.Fatal("unexpected call")
					return nil
				},
			); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("it calls the function again after it fails", func(t *testing.T) {
			var once Once
			want := errors.New("<error>")

			if got := once.Do(
				t.Context(),
				func(context.Context) error {
					return want
				},
			); got != want {
				t.Fatalf("unexpected error: got %v, want %v", got, want)
			}

			called := false
			if err := once.Do(
				t.Context(),
				func(context.Context) error {
					called = true
					return nil
				},
			); err != nil {
				t.Fatal(err)
			}

			if !called {
				t.Fatal("expected function to be called")
			}
		})

		t.Run("it calls the function again after it panics", func(t *testing.T) {
			var once Once

			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected a panic")
					}
				}()

				once.Do(
					t.Context(),
					func(context.Context) error {
						panic("<panic>")
					},
				)
			}()

			called := false
			if err := once.Do(
				t.Context(),
				func(context.Context) error {
					called = true
					return nil
				},
			); err != nil {
				t.Fatal(err)
			}

			if !called {
				t.Fatal("expected function to be called")
			}
		})

		t.Run("it passes the context to the function", func(t *testing.T) {
			type key struct{}

			var once Once
			ctx := context.WithValue(t.Context(), key{}, "<value>")

			if err := once.Do(
				ctx,
				func(ctx context.Context) error {
					if got := ctx.Value(key{}); got != "<value>" {
						t.Errorf("unexpected context value: got %v, want <value>", got)
					}
					return nil
				},
			); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("when the context is done and the function has not succeeded", func(t *testing.T) {
			t.Run("it returns the context error without calling the function", func(t *testing.T) {
				var once Once

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				// Repeat to catch a nondeterministic choice between the two.
				for range 100 {
					err := once.Do(
						ctx,
						func(context.Context) error {
							t.Fatal("unexpected call")
							return nil
						},
					)

					if err != context.Canceled {
						t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
					}
				}
			})
		})

		t.Run("when the context is done and the function has succeeded", func(t *testing.T) {
			t.Run("it returns nil", func(t *testing.T) {
				var once Once

				if err := once.Do(
					t.Context(),
					func(context.Context) error {
						return nil
					},
				); err != nil {
					t.Fatal(err)
				}

				ctx, cancel := context.WithCancel(t.Context())
				cancel()

				if err := once.Do(
					ctx,
					func(context.Context) error {
						t.Fatal("unexpected call")
						return nil
					},
				); err != nil {
					t.Fatal(err)
				}
			})
		})

		t.Run("when another call is in progress", func(t *testing.T) {
			t.Run("it stops waiting if the context is canceled", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var once Once
					release := make(chan struct{})

					go once.Do(
						t.Context(),
						func(context.Context) error {
							<-release
							return nil
						},
					)

					// Block until the first call is running its function.
					synctest.Wait()

					ctx, cancel := context.WithCancel(t.Context())

					var err error
					go func() {
						err = once.Do(
							ctx,
							func(context.Context) error {
								t.Error("unexpected call")
								return nil
							},
						)
					}()

					synctest.Wait()
					cancel()
					synctest.Wait()

					if err != context.Canceled {
						t.Fatalf("unexpected error: got %v, want %v", err, context.Canceled)
					}

					close(release)
					synctest.Wait()

					// The first call succeeded, so the function is not called
					// again.
					if err := once.Do(
						t.Context(),
						func(context.Context) error {
							t.Fatal("unexpected call")
							return nil
						},
					); err != nil {
						t.Fatal(err)
					}
				})
			})

			t.Run("it returns nil once the other call succeeds", func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					var once Once
					release := make(chan struct{})

					go once.Do(
						t.Context(),
						func(context.Context) error {
							<-release
							return nil
						},
					)

					synctest.Wait()

					var err error
					go func() {
						err = once.Do(
							t.Context(),
							func(context.Context) error {
								t.Error("unexpected call")
								return nil
							},
						)
					}()

					synctest.Wait()
					close(release)
					synctest.Wait()

					if err != nil {
						t.Fatal(err)
					}
				})
			})
		})
	})
}

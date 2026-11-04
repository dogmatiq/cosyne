package cosyne_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"testing/synctest"

	. "github.com/dogmatiq/cosyne"
)

func TestObservable(t *testing.T) {
	t.Run("func Load()", func(t *testing.T) {
		t.Run("it returns the zero value before any value is stored", func(t *testing.T) {
			var value Observable[int]

			if v, _ := value.Load(); v != 0 {
				t.Fatalf("unexpected value: got %d, want 0", v)
			}
		})

		t.Run("it returns the most recently stored value", func(t *testing.T) {
			var value Observable[int]
			value.Store(1)
			value.Store(2)

			if v, _ := value.Load(); v != 2 {
				t.Fatalf("unexpected value: got %d, want 2", v)
			}
		})

		t.Run("it returns a channel that is closed when the value is replaced", func(t *testing.T) {
			var value Observable[int]

			// The initial value has a channel too.
			_, changed := value.Load()

			select {
			case <-changed:
				t.Fatal("did not expect channel to be closed before a store")
			default:
			}

			value.Store(1)

			select {
			case <-changed:
			default:
				t.Fatal("expected channel to be closed after a store")
			}

			_, changed = value.Load()
			value.Store(2)

			select {
			case <-changed:
			default:
				t.Fatal("expected channel to be closed after a second store")
			}
		})
	})

	t.Run("it notifies multiple readers of changes", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const readerCount = 10

			var (
				value   Observable[int]
				done    = make(chan struct{})
				want    = []int{0}
				readers sync.WaitGroup
				errs    [readerCount]error
			)

			go func() {
				for range 10 {
					// Block until the readers are waiting for a change.
					synctest.Wait()

					n := rand.Int()
					want = append(want, n)
					value.Store(n)
				}

				close(done)
			}()

			for i := range readerCount {
				readers.Go(func() {
					var got []int

					for {
						v, changed := value.Load()
						got = append(got, v)

						select {
						case <-changed:
						case <-done:
							if !slices.Equal(got, want) {
								errs[i] = fmt.Errorf("unexpected values: got %v, want %v", got, want)
							}
							return
						}
					}
				})
			}

			readers.Wait()

			for _, err := range errs {
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	})

	t.Run("it does not produce unexpected values with multiple writers", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var (
				value   Observable[int]
				stored  sync.Map
				writers sync.WaitGroup
			)

			stored.Store(0, true)

			for range 3 {
				writers.Go(func() {
					for range 100 {
						n := rand.Int()
						stored.Store(n, true)
						value.Store(n)
					}
				})
			}

			done := make(chan struct{})
			go func() {
				writers.Wait()
				close(done)
			}()

			for {
				v, changed := value.Load()

				if _, ok := stored.Load(v); !ok {
					t.Fatalf("unexpected value: got %d", v)
				}

				select {
				case <-changed:
				case <-done:
					return
				}
			}
		})
	})

	t.Run("func ObservableW", func(t *testing.T) {
		t.Run("it stores the value in the observable it views", func(t *testing.T) {
			var value Observable[int]
			var w ObservableW[int] = &value

			w.Store(42)

			if v, _ := value.Load(); v != 42 {
				t.Fatalf("unexpected value: got %d, want 42", v)
			}
		})
	})
}

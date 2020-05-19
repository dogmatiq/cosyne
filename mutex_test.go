package cosyne_test

import (
	"context"
	"time"

	. "github.com/dogmatiq/cosyne"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Mutex", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		mutex  *Mutex
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
		mutex = &Mutex{}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("func Lock()", func() {
		It("blocks calls to Lock()", func() {
			err := mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			err = mutex.Lock(ctx)
			Expect(err).To(Equal(context.DeadlineExceeded))
		})

		It("causes TryLock() to return false", func() {
			err := mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			ok := mutex.TryLock()
			Expect(ok).To(BeFalse())
		})
	})

	Describe("func TryLock()", func() {
		It("blocks calls to Lock()", func() {
			ok := mutex.TryLock()
			Expect(ok).To(BeTrue())

			err := mutex.Lock(ctx)
			Expect(err).To(Equal(context.DeadlineExceeded))
		})

		It("causes TryLock() to return false", func() {
			ok := mutex.TryLock()
			Expect(ok).To(BeTrue())

			ok = mutex.TryLock()
			Expect(ok).To(BeFalse())
		})
	})

	Describe("func Unlock()", func() {
		It("allows subsequent calls to Lock()", func() {
			err := mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			mutex.Unlock()

			err = mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("allows subsequent calls to TryLock()", func() {
			err := mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			mutex.Unlock()

			ok := mutex.TryLock()
			Expect(ok).To(BeTrue())
		})

		It("unblocks one blocking call to Lock()", func() {
			err := mutex.Lock(ctx)
			Expect(err).ShouldNot(HaveOccurred())

			errors := make(chan error, 2)
			fn := func() { errors <- mutex.Lock(ctx) }

			go fn()
			go fn()

			time.Sleep(5 * time.Millisecond)
			mutex.Unlock()

			Expect(<-errors).ShouldNot(HaveOccurred())
			Expect(<-errors).To(Equal(context.DeadlineExceeded))
		})

		It("panics if the mutex is not locked", func() {
			Expect(func() {
				mutex.Unlock()
			}).To(Panic())
		})
	})
})

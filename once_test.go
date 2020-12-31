package cosyne_test

import (
	"context"
	"errors"
	"time"

	. "github.com/dogmatiq/cosyne"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Once", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		once   *Once
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
		once = &Once{}
	})

	AfterEach(func() {
		cancel()
	})

	Describe("func Do()", func() {
		It("does not call the function again after success", func() {
			called := false
			err := once.Do(ctx, func(ctx context.Context) error {
				called = true
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())

			err = once.Do(ctx, func(ctx context.Context) error {
				Fail("unexpected call")
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("calls the function again after an error returns", func() {
			err := once.Do(ctx, func(ctx context.Context) error {
				return errors.New("<error>")
			})
			Expect(err).To(MatchError("<error>"))

			called := false
			err = once.Do(ctx, func(ctx context.Context) error {
				called = true
				return nil
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})

		It("blocks conccurent calls to Do()", func() {
			otherCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			barrier := make(chan struct{})
			go once.Do(otherCtx, func(ctx context.Context) error {
				close(barrier)
				<-ctx.Done()
				return nil
			})

			<-barrier
			err := once.Do(ctx, func(context.Context) error {
				Fail("unexpected call")
				return nil
			})
			Expect(err).To(Equal(context.DeadlineExceeded))
		})
	})
})

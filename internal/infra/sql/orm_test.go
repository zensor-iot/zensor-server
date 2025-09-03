package sql_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ORM Timeout", func() {
	var (
		orm sql.ORM
		ctx context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ctx = context.Background()
	})

	ginkgo.Context("WithTimeout", func() {
		ginkgo.When("creating a timeout context", func() {
			ginkgo.It("should create a context with timeout", func() {
				timeout := 5 * time.Second
				timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Verify the context has a deadline
				deadline, ok := timeoutCtx.Deadline()
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(deadline).To(gomega.BeTemporally("~", time.Now().Add(timeout), time.Second))
			})

			ginkgo.It("should work with WithTimeout method", func() {
				timeout := 2 * time.Second
				timeoutORM := orm.WithTimeout(ctx, timeout)

				// Verify the ORM instance is returned
				gomega.Expect(timeoutORM).NotTo(gomega.BeNil())
			})
		})

		ginkgo.When("using timeout context with database operations", func() {
			ginkgo.It("should complete operations within timeout", func() {
				timeout := 5 * time.Second
				timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Create a simple table for testing
				type TestModel struct {
					ID   uint `gorm:"primaryKey"`
					Name string
				}

				err := orm.AutoMigrate(&TestModel{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Test that operations work with timeout context
				var count int64
				err = orm.WithContext(timeoutCtx).Model(&TestModel{}).Count(&count).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(count).To(gomega.Equal(int64(0)))
			})
		})
	})

	ginkgo.Context("Default Timeout", func() {
		ginkgo.When("using WithContext with default timeout", func() {
			ginkgo.It("should work with memory ORM (no timeout)", func() {
				// Memory ORM has timeout = 0, so WithContext should work normally
				type TestModel struct {
					ID   uint `gorm:"primaryKey"`
					Name string
				}

				err := orm.AutoMigrate(&TestModel{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				var count int64
				err = orm.WithContext(ctx).Model(&TestModel{}).Count(&count).Error()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(count).To(gomega.Equal(int64(0)))
			})
		})
	})

	ginkgo.Context("Context WithTimeout", func() {
		ginkgo.When("creating timeout context", func() {
			ginkgo.It("should return context and cancel function", func() {
				timeout := 10 * time.Second
				timeoutCtx, cancel := context.WithTimeout(ctx, timeout)

				// Verify we get both context and cancel function
				gomega.Expect(timeoutCtx).NotTo(gomega.BeNil())
				gomega.Expect(cancel).NotTo(gomega.BeNil())

				// Verify the context has a deadline
				deadline, ok := timeoutCtx.Deadline()
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(deadline).To(gomega.BeTemporally("~", time.Now().Add(timeout), time.Second))

				// Test that cancel function works
				cancel()
				select {
				case <-timeoutCtx.Done():
					// Context should be cancelled
				case <-time.After(100 * time.Millisecond):
					ginkgo.Fail("Context should be cancelled")
				}
			})

			ginkgo.It("should handle zero timeout", func() {
				timeoutCtx, cancel := context.WithTimeout(ctx, 0)
				defer cancel()

				// With zero timeout, context should be cancelled immediately
				select {
				case <-timeoutCtx.Done():
					// Context should be cancelled
				case <-time.After(100 * time.Millisecond):
					ginkgo.Fail("Context should be cancelled immediately with zero timeout")
				}
			})
		})
	})
})

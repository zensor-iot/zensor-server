package domain_test

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("AllowedUser", func() {
	ginkgo.Context("AllowedUserBuilder", func() {
		ginkgo.When("building with a valid email", func() {
			ginkgo.It("should build with generated ID and normalized email", func() {
				user, err := domain.NewAllowedUserBuilder().
					WithEmail("Sebastian@Example.COM").
					WithDisplayName("Sebastian").
					WithIsAdmin(true).
					Build()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(user.ID).NotTo(gomega.BeEmpty())
				gomega.Expect(user.Email).To(gomega.Equal("sebastian@example.com"))
				gomega.Expect(user.DisplayName).To(gomega.Equal("Sebastian"))
				gomega.Expect(user.IsAdmin).To(gomega.BeTrue())
				gomega.Expect(user.CreatedAt).NotTo(gomega.BeZero())
				gomega.Expect(user.UpdatedAt).NotTo(gomega.BeZero())
				gomega.Expect(user.LastLoginAt).To(gomega.BeNil())
			})
		})

		ginkgo.When("building with an invalid email", func() {
			ginkgo.It("should return an error", func() {
				_, err := domain.NewAllowedUserBuilder().
					WithEmail("not-an-email").
					Build()

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})

		ginkgo.When("building without an email", func() {
			ginkgo.It("should return an error", func() {
				_, err := domain.NewAllowedUserBuilder().Build()

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("RecordLogin", func() {
		ginkgo.When("a user logs in", func() {
			ginkgo.It("should update display name and last login timestamp", func() {
				user, err := domain.NewAllowedUserBuilder().
					WithEmail("user@example.com").
					Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				user.RecordLogin("New Name")

				gomega.Expect(user.DisplayName).To(gomega.Equal("New Name"))
				gomega.Expect(user.LastLoginAt).NotTo(gomega.BeNil())
			})
		})
	})
})

var _ = ginkgo.Describe("Session", func() {
	ginkgo.Context("IsExpired", func() {
		ginkgo.When("the session expiry is in the future", func() {
			ginkgo.It("should not be expired", func() {
				session := domain.Session{ExpiresAt: time.Now().Add(time.Hour)}

				gomega.Expect(session.IsExpired(time.Now())).To(gomega.BeFalse())
			})
		})

		ginkgo.When("the session expiry is in the past", func() {
			ginkgo.It("should be expired", func() {
				session := domain.Session{ExpiresAt: time.Now().Add(-time.Hour)}

				gomega.Expect(session.IsExpired(time.Now())).To(gomega.BeTrue())
			})
		})
	})
})

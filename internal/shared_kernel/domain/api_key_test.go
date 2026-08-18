package domain_test

import (
	"crypto/sha256"
	"encoding/hex"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("APIKey", func() {
	ginkgo.Context("Build", func() {
		ginkgo.When("all required fields are provided", func() {
			var (
				apiKey    domain.APIKey
				plaintext string
				err       error
			)

			ginkgo.BeforeEach(func() {
				apiKey, plaintext, err = domain.NewAPIKeyBuilder().
					WithName("grafana-sync").
					WithCreatedBy(domain.ID("admin-1")).
					Build()
			})

			ginkgo.It("should not return an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should generate a plaintext key with the zsk_ prefix and 32 random bytes hex-encoded", func() {
				gomega.Expect(plaintext).To(gomega.HavePrefix("zsk_"))
				gomega.Expect(plaintext).To(gomega.HaveLen(68))
			})

			ginkgo.It("should store the SHA-256 hex digest of the plaintext key", func() {
				digest := sha256.Sum256([]byte(plaintext))
				gomega.Expect(apiKey.KeyHash).To(gomega.Equal(hex.EncodeToString(digest[:])))
			})

			ginkgo.It("should store the first 12 characters of the plaintext as prefix", func() {
				gomega.Expect(apiKey.KeyPrefix).To(gomega.Equal(plaintext[:12]))
			})

			ginkgo.It("should populate identity fields", func() {
				gomega.Expect(apiKey.ID).NotTo(gomega.BeEmpty())
				gomega.Expect(apiKey.Name).To(gomega.Equal("grafana-sync"))
				gomega.Expect(apiKey.CreatedBy).To(gomega.Equal(domain.ID("admin-1")))
				gomega.Expect(apiKey.CreatedAt).NotTo(gomega.BeZero())
			})

			ginkgo.It("should generate a distinct key on every build", func() {
				_, otherPlaintext, otherErr := domain.NewAPIKeyBuilder().
					WithName("other").
					WithCreatedBy(domain.ID("admin-1")).
					Build()
				gomega.Expect(otherErr).NotTo(gomega.HaveOccurred())
				gomega.Expect(otherPlaintext).NotTo(gomega.Equal(plaintext))
			})
		})

		ginkgo.When("the name is blank", func() {
			ginkgo.It("should return an error", func() {
				_, _, err := domain.NewAPIKeyBuilder().
					WithName("   ").
					WithCreatedBy(domain.ID("admin-1")).
					Build()
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})

		ginkgo.When("the name is missing", func() {
			ginkgo.It("should return an error", func() {
				_, _, err := domain.NewAPIKeyBuilder().
					WithCreatedBy(domain.ID("admin-1")).
					Build()
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})
})

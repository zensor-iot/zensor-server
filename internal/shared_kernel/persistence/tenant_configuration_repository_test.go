package persistence_test

import (
	"time"
	"zensor-server/internal/shared_kernel/persistence/internal"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("TenantConfigurationRepository", func() {
	ginkgo.Context("Create", func() {
		var config domain.TenantConfiguration

		ginkgo.When("creating a tenant configuration", func() {
			ginkgo.BeforeEach(func() {
				config = domain.TenantConfiguration{
					ID:        "test-id",
					TenantID:  "tenant-id",
					Timezone:  "UTC",
					Version:   1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			})

			ginkgo.It("should convert internal model correctly", func() {
				// Test the internal model conversion
				internalConfig := internal.FromTenantConfiguration(config)
				domainConfig := internalConfig.ToDomain()

				gomega.Expect(domainConfig.ID).To(gomega.Equal(config.ID))
				gomega.Expect(domainConfig.TenantID).To(gomega.Equal(config.TenantID))
				gomega.Expect(domainConfig.Timezone).To(gomega.Equal(config.Timezone))
			})
		})
	})

	ginkgo.Context("TenantConfigurationBuilder", func() {
		var (
			tenantID string
			timezone string
		)

		ginkgo.When("building a tenant configuration", func() {
			ginkgo.BeforeEach(func() {
				tenantID = "test-tenant-id"
				timezone = "America/New_York"
			})

			ginkgo.It("should build configuration successfully", func() {
				config, err := domain.NewTenantConfigurationBuilder().
					WithTenantID(domain.ID(tenantID)).
					WithTimezone(timezone).
					Build()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(config.TenantID).To(gomega.Equal(domain.ID(tenantID)))
				gomega.Expect(config.Timezone).To(gomega.Equal(timezone))
				gomega.Expect(config.ID).NotTo(gomega.BeEmpty())
				gomega.Expect(config.Version).To(gomega.Equal(1))
			})
		})
	})
})

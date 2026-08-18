package persistence_test

import (
	"context"

	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("EvaluationRuleRepository", func() {
	var (
		orm  sql.ORM
		repo usecases.EvaluationRuleRepository
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		repo, err = persistence.NewEvaluationRuleRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(repo).NotTo(gomega.BeNil())

		ctx = context.Background()
	})

	ginkgo.Context("AddToDevice", func() {
		var device domain.Device
		var rule domain.EvaluationRule

		ginkgo.BeforeEach(func() {
			device = domain.Device{
				ID: domain.ID(utils.GenerateUUID()),
			}
			rule = domain.EvaluationRule{
				ID:          domain.ID(utils.GenerateUUID()),
				Description: "test rule",
				Kind:        "threshold",
				Enabled:     true,
				Parameters: []domain.EvaluationRuleParameter{
					{Key: "threshold", Value: float64(42)},
				},
			}
		})

		ginkgo.It("should create the evaluation rule in the database, associated with the device", func() {
			err := repo.AddToDevice(ctx, device, rule)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			result, err := repo.FindAllByDeviceID(ctx, device.ID.String())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result).To(gomega.HaveLen(1))
			gomega.Expect(result[0].ID).To(gomega.Equal(rule.ID))
			gomega.Expect(result[0].Description).To(gomega.Equal(rule.Description))
			gomega.Expect(result[0].Kind).To(gomega.Equal(rule.Kind))
		})
	})

	ginkgo.Context("FindAllByDeviceID", func() {
		ginkgo.When("no evaluation rules exist for the device", func() {
			ginkgo.It("should return an empty list", func() {
				result, err := repo.FindAllByDeviceID(ctx, utils.GenerateUUID())
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeEmpty())
			})
		})
	})
})

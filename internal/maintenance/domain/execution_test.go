package domain_test

import (
	"time"

	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Execution", func() {
	Context("MarkCompleted", func() {
		var execution maintenanceDomain.Execution

		BeforeEach(func() {
			var err error
			execution, err = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(shareddomain.ID(utils.GenerateUUID())).
				WithScheduledDate(time.Now().AddDate(0, 0, -1)).
				WithFieldValues(map[string]any{"filter_type": "carbon", "notes": "prefilled"}).
				Build()
			Expect(err).NotTo(HaveOccurred())
		})

		When("field values are captured on completion", func() {
			It("should record completion and merge captured values over defaults", func() {
				execution.MarkCompleted("user@example.com", map[string]any{"notes": "replaced", "hours": 2})

				Expect(execution.IsCompleted()).To(BeTrue())
				Expect(string(*execution.CompletedBy)).To(Equal("user@example.com"))
				Expect(execution.FieldValues).To(HaveKeyWithValue("filter_type", "carbon"))
				Expect(execution.FieldValues).To(HaveKeyWithValue("notes", "replaced"))
				Expect(execution.FieldValues).To(HaveKeyWithValue("hours", 2))
			})
		})

		When("no field values are captured", func() {
			It("should keep the existing field values", func() {
				execution.MarkCompleted("user@example.com", nil)

				Expect(execution.IsCompleted()).To(BeTrue())
				Expect(execution.FieldValues).To(HaveKeyWithValue("filter_type", "carbon"))
				Expect(execution.FieldValues).To(HaveKeyWithValue("notes", "prefilled"))
			})
		})
	})
})

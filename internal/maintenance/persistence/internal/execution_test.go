package internal_test

import (
	"encoding/json"
	"time"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	persistenceInternal "zensor-server/internal/maintenance/persistence/internal"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MaintenanceExecution Internal Model", func() {
	ginkgo.Context("FieldValues Value/Scan", func() {
		ginkgo.When("serializing FieldValues to driver.Value", func() {
			ginkgo.It("should return empty object string for empty map", func() {
				fieldValues := persistenceInternal.FieldValues{}
				value, err := fieldValues.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal("{}"))
			})

			ginkgo.It("should return JSON marshaled bytes for non-empty map", func() {
				fieldValues := persistenceInternal.FieldValues{
					"maintenance_type": "Filter Replacement",
					"provider":         "ACME Plumbing",
					"cost":             float64(150.50),
				}
				value, err := fieldValues.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				bytes, ok := value.([]byte)
				gomega.Expect(ok).To(gomega.BeTrue())

				var result map[string]any
				err = json.Unmarshal(bytes, &result)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result["maintenance_type"]).To(gomega.Equal("Filter Replacement"))
				gomega.Expect(result["cost"]).To(gomega.BeNumerically("~", 150.50, 0.01))
			})
		})

		ginkgo.When("deserializing driver.Value to FieldValues", func() {
			ginkgo.It("should handle string input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan(`{"key":"value"}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fieldValues).To(gomega.HaveKey("key"))
				gomega.Expect(fieldValues["key"]).To(gomega.Equal("value"))
			})

			ginkgo.It("should handle byte slice input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan([]byte(`{"field":"test","number":123}`))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fieldValues).To(gomega.HaveKey("field"))
				gomega.Expect(fieldValues["field"]).To(gomega.Equal("test"))
				gomega.Expect(fieldValues["number"]).To(gomega.BeNumerically("==", 123))
			})

			ginkgo.It("should handle empty string input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan("")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fieldValues).To(gomega.BeEmpty())
			})

			ginkgo.It("should handle empty byte slice input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan([]byte{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fieldValues).To(gomega.BeEmpty())
			})

			ginkgo.It("should handle nil input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan(nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fieldValues).To(gomega.BeEmpty())
			})

			ginkgo.It("should return error for invalid input", func() {
				var fieldValues persistenceInternal.FieldValues
				err := fieldValues.Scan(123)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid type for field_values"))
			})
		})
	})

	ginkgo.Context("ToDomain", func() {
		var internalExecution persistenceInternal.Execution

		ginkgo.BeforeEach(func() {
			now := utils.Time{Time: time.Now()}
			internalExecution = persistenceInternal.Execution{
				ID:            "execution-id",
				Version:       1,
				ActivityID:    "activity-id",
				ScheduledDate: now,
				CompletedAt:   nil,
				CompletedBy:   nil,
				OverdueDays:   0,
				FieldValues:   persistenceInternal.FieldValues{},
				CreatedAt:     now,
				UpdatedAt:     now,
				DeletedAt:     nil,
			}
		})

		ginkgo.When("converting to domain", func() {
			ginkgo.It("should convert basic fields correctly", func() {
				domain := internalExecution.ToDomain()
				gomega.Expect(domain.ID).To(gomega.Equal(shareddomain.ID("execution-id")))
				gomega.Expect(domain.Version).To(gomega.Equal(shareddomain.Version(1)))
				gomega.Expect(domain.ActivityID).To(gomega.Equal(shareddomain.ID("activity-id")))
				gomega.Expect(domain.ScheduledDate).To(gomega.Equal(internalExecution.ScheduledDate))
				gomega.Expect(domain.OverdueDays).To(gomega.Equal(maintenanceDomain.OverdueDays(0)))
				gomega.Expect(domain.FieldValues).To(gomega.BeEmpty())
			})

			ginkgo.It("should handle CompletedAt", func() {
				now := utils.Time{Time: time.Now()}
				internalExecution.CompletedAt = &now
				domain := internalExecution.ToDomain()
				gomega.Expect(domain.CompletedAt).NotTo(gomega.BeNil())
			})

			ginkgo.It("should handle CompletedBy", func() {
				completedBy := "user@example.com"
				internalExecution.CompletedBy = &completedBy
				domain := internalExecution.ToDomain()
				gomega.Expect(domain.CompletedBy).NotTo(gomega.BeNil())
				gomega.Expect(string(*domain.CompletedBy)).To(gomega.Equal("user@example.com"))
			})

			ginkgo.It("should handle DeletedAt", func() {
				now := utils.Time{Time: time.Now()}
				internalExecution.DeletedAt = &now
				domain := internalExecution.ToDomain()
				gomega.Expect(domain.DeletedAt).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("FromMaintenanceExecution", func() {
		var domainExecution maintenanceDomain.Execution

		ginkgo.BeforeEach(func() {
			domainExecution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(shareddomain.ID("activity-id")).
				WithScheduledDate(time.Now()).
				WithFieldValues(map[string]any{
					"maintenance_type": "Filter Replacement",
					"cost":             float64(150.50),
				}).
				Build()
		})

		ginkgo.When("converting from domain", func() {
			ginkgo.It("should convert basic fields correctly", func() {
				internal := persistenceInternal.FromExecution(domainExecution)
				gomega.Expect(internal.ID).To(gomega.Equal(domainExecution.ID.String()))
				gomega.Expect(internal.Version).To(gomega.Equal(int(domainExecution.Version)))
				gomega.Expect(internal.ActivityID).To(gomega.Equal(domainExecution.ActivityID.String()))
				gomega.Expect(internal.ScheduledDate).To(gomega.Equal(domainExecution.ScheduledDate))
				gomega.Expect(internal.OverdueDays).To(gomega.Equal(int(domainExecution.OverdueDays)))
				gomega.Expect(internal.FieldValues).To(gomega.HaveKey("maintenance_type"))
				gomega.Expect(internal.FieldValues["maintenance_type"]).To(gomega.Equal("Filter Replacement"))
			})

			ginkgo.It("should handle CompletedAt", func() {
				now := utils.Time{Time: time.Now()}
				domainExecution.CompletedAt = &now
				internal := persistenceInternal.FromExecution(domainExecution)
				gomega.Expect(internal.CompletedAt).NotTo(gomega.BeNil())
			})

			ginkgo.It("should handle CompletedBy", func() {
				completedBy := maintenanceDomain.CompletedBy("user@example.com")
				domainExecution.CompletedBy = &completedBy
				internal := persistenceInternal.FromExecution(domainExecution)
				gomega.Expect(internal.CompletedBy).NotTo(gomega.BeNil())
				gomega.Expect(*internal.CompletedBy).To(gomega.Equal("user@example.com"))
			})

			ginkgo.It("should handle DeletedAt", func() {
				now := utils.Time{Time: time.Now()}
				domainExecution.DeletedAt = &now
				internal := persistenceInternal.FromExecution(domainExecution)
				gomega.Expect(internal.DeletedAt).NotTo(gomega.BeNil())
			})
		})
	})
})

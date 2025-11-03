package internal_test

import (
	"time"
	maintenance_httpapi_internal "zensor-server/internal/maintenance/httpapi/internal"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/infra/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MaintenanceExecution", func() {
	var execution maintenanceDomain.MaintenanceExecution
	var completedBy string

	BeforeEach(func() {
		completedBy = "user@example.com"
		execution, _ = maintenanceDomain.NewMaintenanceExecutionBuilder().
			WithActivityID(shareddomain.ID("activity-123")).
			WithScheduledDate(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)).
			WithFieldValues(map[string]any{
				"maintenance_type": "Filter Replacement",
				"cost":             150.50,
			}).
			Build()

		execution.ID = shareddomain.ID("execution-123")
		execution.Version = 3
		execution.CreatedAt = utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
		execution.UpdatedAt = utils.Time{Time: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)}
	})

	Context("ToMaintenanceExecutionResponse", func() {
		var response maintenance_httpapi_internal.MaintenanceExecutionResponse

		BeforeEach(func() {
			response = maintenance_httpapi_internal.ToMaintenanceExecutionResponse(execution)
		})

		When("converting execution with all fields", func() {
			It("should map ID correctly", func() {
				Expect(response.ID).To(Equal("execution-123"))
			})

			It("should map Version correctly", func() {
				Expect(response.Version).To(Equal(3))
			})

			It("should map ActivityID correctly", func() {
				Expect(response.ActivityID).To(Equal("activity-123"))
			})

			It("should map ScheduledDate correctly", func() {
				Expect(response.ScheduledDate).To(Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)))
			})

			It("should map CreatedAt correctly", func() {
				Expect(response.CreatedAt).To(Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
			})

			It("should map UpdatedAt correctly", func() {
				Expect(response.UpdatedAt).To(Equal(time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)))
			})
		})

		When("converting field values", func() {
			It("should map field values correctly", func() {
				Expect(response.FieldValues).To(HaveKey("maintenance_type"))
				Expect(response.FieldValues).To(HaveKey("cost"))
				Expect(response.FieldValues["maintenance_type"]).To(Equal("Filter Replacement"))
				Expect(response.FieldValues["cost"]).To(Equal(150.50))
			})
		})

		When("execution is not overdue", func() {
			BeforeEach(func() {
				execution.ScheduledDate = utils.Time{Time: time.Now().AddDate(0, 0, 30)}
				execution.OverdueDays = 0
				response = maintenance_httpapi_internal.ToMaintenanceExecutionResponse(execution)
			})

			It("should map IsOverdue correctly", func() {
				Expect(response.IsOverdue).To(BeFalse())
			})

			It("should map OverdueDays correctly", func() {
				Expect(response.OverdueDays).To(Equal(0))
			})
		})

		When("execution is overdue", func() {
			BeforeEach(func() {
				execution.ScheduledDate = utils.Time{Time: time.Now().AddDate(0, 0, -15)}
				execution.OverdueDays = 15
				response = maintenance_httpapi_internal.ToMaintenanceExecutionResponse(execution)
			})

			It("should map IsOverdue correctly", func() {
				Expect(response.IsOverdue).To(BeTrue())
			})

			It("should map OverdueDays correctly", func() {
				Expect(response.OverdueDays).To(Equal(15))
			})
		})

		When("execution is not completed", func() {
			It("should set CompletedAt to nil", func() {
				Expect(response.CompletedAt).To(BeNil())
			})

			It("should set CompletedBy to nil", func() {
				Expect(response.CompletedBy).To(BeNil())
			})
		})

		When("execution is completed", func() {
			BeforeEach(func() {
				completedAt := utils.Time{Time: time.Date(2024, 2, 1, 10, 30, 0, 0, time.UTC)}
				execution.CompletedAt = &completedAt
				completedByVO := maintenanceDomain.CompletedBy(completedBy)
				execution.CompletedBy = &completedByVO
				response = maintenance_httpapi_internal.ToMaintenanceExecutionResponse(execution)
			})

			It("should map CompletedAt correctly", func() {
				Expect(response.CompletedAt).NotTo(BeNil())
				expectedTime := time.Date(2024, 2, 1, 10, 30, 0, 0, time.UTC)
				Expect(*response.CompletedAt).To(Equal(expectedTime))
			})

			It("should map CompletedBy correctly", func() {
				Expect(response.CompletedBy).NotTo(BeNil())
				Expect(*response.CompletedBy).To(Equal(completedBy))
			})
		})
	})

	Context("ToMaintenanceExecutionListResponse", func() {
		var executions []maintenanceDomain.MaintenanceExecution
		var response maintenance_httpapi_internal.MaintenanceExecutionListResponse

		BeforeEach(func() {
			execution1, _ := maintenanceDomain.NewMaintenanceExecutionBuilder().
				WithActivityID(shareddomain.ID("activity-123")).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"key": "value1"}).
				Build()

			execution2, _ := maintenanceDomain.NewMaintenanceExecutionBuilder().
				WithActivityID(shareddomain.ID("activity-123")).
				WithScheduledDate(time.Now().AddDate(0, 0, 60)).
				WithFieldValues(map[string]any{"key": "value2"}).
				Build()

			executions = []maintenanceDomain.MaintenanceExecution{execution1, execution2}
		})

		When("converting list of executions", func() {
			BeforeEach(func() {
				response = maintenance_httpapi_internal.ToMaintenanceExecutionListResponse(executions)
			})

			It("should convert all executions", func() {
				Expect(response.Data).To(HaveLen(2))
			})

			It("should map first execution correctly", func() {
				execution1 := response.Data[0]
				Expect(execution1.ActivityID).To(Equal("activity-123"))
				Expect(execution1.FieldValues).To(HaveKey("key"))
				Expect(execution1.FieldValues["key"]).To(Equal("value1"))
			})

			It("should map second execution correctly", func() {
				execution2 := response.Data[1]
				Expect(execution2.ActivityID).To(Equal("activity-123"))
				Expect(execution2.FieldValues).To(HaveKey("key"))
				Expect(execution2.FieldValues["key"]).To(Equal("value2"))
			})
		})

		When("converting empty list", func() {
			BeforeEach(func() {
				executions = []maintenanceDomain.MaintenanceExecution{}
				response = maintenance_httpapi_internal.ToMaintenanceExecutionListResponse(executions)
			})

			It("should return empty data array", func() {
				Expect(response.Data).To(HaveLen(0))
			})
		})
	})
})


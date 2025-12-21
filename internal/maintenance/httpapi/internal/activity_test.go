package internal_test

import (
	"time"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenance_httpapi_internal "zensor-server/internal/maintenance/httpapi/internal"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MaintenanceActivity", func() {
	var activity maintenanceDomain.Activity
	var customTypeName string

	BeforeEach(func() {
		customTypeName = "CustomType"
		activityType, _ := maintenanceDomain.NewActivityTypeBuilder().
			WithName(maintenanceDomain.ActivityTypeWaterSystem).
			WithDisplayName("Water System Maintenance").
			WithDescription("Water system maintenance tasks").
			WithIsPredefined(true).
			Build()

		field1 := maintenanceDomain.FieldDefinition{
			Name:        shareddomain.Name("field1"),
			DisplayName: shareddomain.DisplayName("Field 1"),
			Type:        maintenanceDomain.FieldTypeText,
			IsRequired:  true,
		}

		defaultValue := interface{}("default")
		field2 := maintenanceDomain.FieldDefinition{
			Name:         shareddomain.Name("field2"),
			DisplayName:  shareddomain.DisplayName("Field 2"),
			Type:         maintenanceDomain.FieldTypeText,
			IsRequired:   false,
			DefaultValue: &defaultValue,
		}

		activity, _ = maintenanceDomain.NewActivityBuilder().
			WithTenantID(shareddomain.ID("tenant-123")).
			WithType(activityType).
			WithCustomTypeName(customTypeName).
			WithName("Test Activity").
			WithDescription("Test Description").
			WithSchedule("0 0 1 * *").
			WithNotificationDaysBefore([]int{7, 3, 1}).
			WithFields([]maintenanceDomain.FieldDefinition{field1, field2}).
			Build()

		activity.ID = shareddomain.ID("activity-123")
		activity.Version = 5
		activity.IsActive = true
		activity.CreatedAt = utils.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
		activity.UpdatedAt = utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}
	})

	Context("ToMaintenanceActivityResponse", func() {
		var response maintenance_httpapi_internal.ActivityResponse

		BeforeEach(func() {
			response = maintenance_httpapi_internal.ToActivityResponse(activity)
		})

		When("converting activity with all fields", func() {
			It("should map ID correctly", func() {
				Expect(response.ID).To(Equal("activity-123"))
			})

			It("should map Version correctly", func() {
				Expect(response.Version).To(Equal(5))
			})

			It("should map TenantID correctly", func() {
				Expect(response.TenantID).To(Equal("tenant-123"))
			})

			It("should map TypeName correctly", func() {
				Expect(response.TypeName).To(Equal(maintenanceDomain.ActivityTypeWaterSystem))
			})

			It("should map CustomTypeName correctly", func() {
				Expect(response.CustomTypeName).NotTo(BeNil())
				Expect(*response.CustomTypeName).To(Equal(customTypeName))
			})

			It("should map Name correctly", func() {
				Expect(response.Name).To(Equal("Test Activity"))
			})

			It("should map Description correctly", func() {
				Expect(response.Description).To(Equal("Test Description"))
			})

			It("should map Schedule correctly", func() {
				Expect(response.Schedule).To(Equal("0 0 1 * *"))
			})

			It("should map NotificationDaysBefore correctly", func() {
				Expect(response.NotificationDaysBefore).To(Equal([]int{7, 3, 1}))
			})

			It("should map IsActive correctly", func() {
				Expect(response.IsActive).To(BeTrue())
			})

			It("should map CreatedAt correctly", func() {
				Expect(response.CreatedAt).To(Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
			})

			It("should map UpdatedAt correctly", func() {
				Expect(response.UpdatedAt).To(Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
			})
		})

		When("converting fields", func() {
			It("should convert all fields", func() {
				Expect(response.Fields).To(HaveLen(2))
			})

			It("should convert field1 correctly", func() {
				field1 := response.Fields[0]
				Expect(field1.Name).To(Equal("field1"))
				Expect(field1.DisplayName).To(Equal("Field 1"))
				Expect(field1.Type).To(Equal(string(maintenanceDomain.FieldTypeText)))
				Expect(field1.IsRequired).To(BeTrue())
				Expect(field1.DefaultValue).To(BeNil())
			})

			It("should convert field2 with default value correctly", func() {
				field2 := response.Fields[1]
				Expect(field2.Name).To(Equal("field2"))
				Expect(field2.DisplayName).To(Equal("Field 2"))
				Expect(field2.Type).To(Equal(string(maintenanceDomain.FieldTypeText)))
				Expect(field2.IsRequired).To(BeFalse())
				Expect(field2.DefaultValue).NotTo(BeNil())
				Expect(*field2.DefaultValue).To(Equal("default"))
			})
		})

		When("activity has no custom type name", func() {
			BeforeEach(func() {
				activity.CustomTypeName = nil
				response = maintenance_httpapi_internal.ToActivityResponse(activity)
			})

			It("should set CustomTypeName to nil", func() {
				Expect(response.CustomTypeName).To(BeNil())
			})
		})

		When("activity has no fields", func() {
			BeforeEach(func() {
				activity.Fields = []maintenanceDomain.FieldDefinition{}
				response = maintenance_httpapi_internal.ToActivityResponse(activity)
			})

			It("should return empty fields array", func() {
				Expect(response.Fields).To(HaveLen(0))
			})
		})

		When("field has non-string default value", func() {
			BeforeEach(func() {
				defaultValue := interface{}(123)
				field := maintenanceDomain.FieldDefinition{
					Name:         shareddomain.Name("field3"),
					DisplayName:  shareddomain.DisplayName("Field 3"),
					Type:         maintenanceDomain.FieldTypeNumber,
					IsRequired:   false,
					DefaultValue: &defaultValue,
				}
				activity.Fields = []maintenanceDomain.FieldDefinition{field}
				response = maintenance_httpapi_internal.ToActivityResponse(activity)
			})

			It("should not set DefaultValue for non-string types", func() {
				Expect(response.Fields[0].DefaultValue).To(BeNil())
			})
		})
	})

	Context("ToMaintenanceActivityListResponse", func() {
		var activities []maintenanceDomain.Activity
		var response maintenance_httpapi_internal.ActivityListResponse

		BeforeEach(func() {
			activityType, _ := maintenanceDomain.NewActivityTypeBuilder().
				WithName(maintenanceDomain.ActivityTypeWaterSystem).
				WithDisplayName("Water System Maintenance").
				WithDescription("Water system maintenance tasks").
				WithIsPredefined(true).
				Build()

			activity1, _ := maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID("tenant-123")).
				WithType(activityType).
				WithName("Activity 1").
				WithDescription("Description 1").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()

			activity2, _ := maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID("tenant-123")).
				WithType(activityType).
				WithName("Activity 2").
				WithDescription("Description 2").
				WithSchedule("0 0 15 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()

			activities = []maintenanceDomain.Activity{activity1, activity2}
		})

		When("converting list of activities", func() {
			BeforeEach(func() {
				response = maintenance_httpapi_internal.ToActivityListResponse(activities)
			})

			It("should convert all activities", func() {
				Expect(response.Data).To(HaveLen(2))
			})

			It("should map first activity correctly", func() {
				activity1 := response.Data[0]
				Expect(activity1.Name).To(Equal("Activity 1"))
				Expect(activity1.Description).To(Equal("Description 1"))
			})

			It("should map second activity correctly", func() {
				activity2 := response.Data[1]
				Expect(activity2.Name).To(Equal("Activity 2"))
				Expect(activity2.Description).To(Equal("Description 2"))
			})
		})

		When("converting empty list", func() {
			BeforeEach(func() {
				activities = []maintenanceDomain.Activity{}
				response = maintenance_httpapi_internal.ToActivityListResponse(activities)
			})

			It("should return empty data array", func() {
				Expect(response.Data).To(HaveLen(0))
			})
		})
	})
})

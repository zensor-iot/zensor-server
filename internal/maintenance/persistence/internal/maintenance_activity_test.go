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

var _ = ginkgo.Describe("MaintenanceActivity Internal Model", func() {
	ginkgo.Context("Days Value/Scan", func() {
		ginkgo.When("serializing Days to driver.Value", func() {
			ginkgo.It("should return empty array string for empty slice", func() {
				days := persistenceInternal.Days{}
				value, err := days.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal("[]"))
			})

			ginkgo.It("should return JSON marshaled bytes for non-empty slice", func() {
				days := persistenceInternal.Days{7, 3, 1}
				value, err := days.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				bytes, ok := value.([]byte)
				gomega.Expect(ok).To(gomega.BeTrue())

				var result []int
				err = json.Unmarshal(bytes, &result)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.Equal([]int{7, 3, 1}))
			})
		})

		ginkgo.When("deserializing driver.Value to Days", func() {
			ginkgo.It("should handle string input", func() {
				var days persistenceInternal.Days
				err := days.Scan(`[7,3,1]`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(days).To(gomega.Equal(persistenceInternal.Days{7, 3, 1}))
			})

			ginkgo.It("should handle byte slice input", func() {
				var days persistenceInternal.Days
				err := days.Scan([]byte(`[5,2]`))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(days).To(gomega.Equal(persistenceInternal.Days{5, 2}))
			})

			ginkgo.It("should handle nil input", func() {
				var days persistenceInternal.Days
				err := days.Scan(nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(days).To(gomega.BeEmpty())
			})

			ginkgo.It("should return error for invalid input", func() {
				var days persistenceInternal.Days
				err := days.Scan(123)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid type for days"))
			})

			ginkgo.It("should handle empty array string", func() {
				var days persistenceInternal.Days
				err := days.Scan(`[]`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(days).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("Fields Value/Scan", func() {
		ginkgo.When("serializing Fields to driver.Value", func() {
			ginkgo.It("should return empty array string for empty slice", func() {
				fields := persistenceInternal.Fields{}
				value, err := fields.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(value).To(gomega.Equal("[]"))
			})

			ginkgo.It("should return JSON marshaled bytes for non-empty slice", func() {
				fields := persistenceInternal.Fields{
					{Name: shareddomain.Name("field1"), DisplayName: shareddomain.DisplayName("Field 1"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
					{Name: shareddomain.Name("field2"), DisplayName: shareddomain.DisplayName("Field 2"), Type: maintenanceDomain.FieldTypeNumber, IsRequired: false},
				}
				value, err := fields.Value()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				bytes, ok := value.([]byte)
				gomega.Expect(ok).To(gomega.BeTrue())

				var result []maintenanceDomain.FieldDefinition
				err = json.Unmarshal(bytes, &result)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.HaveLen(2))
			})
		})

		ginkgo.When("deserializing driver.Value to Fields", func() {
			ginkgo.It("should handle string input", func() {
				var fields persistenceInternal.Fields
				err := fields.Scan(`[{"name":"test","display_name":"Test","type":"text","is_required":true}]`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fields).To(gomega.HaveLen(1))
			})

			ginkgo.It("should handle byte slice input", func() {
				var fields persistenceInternal.Fields
				err := fields.Scan([]byte(`[{"name":"test","display_name":"Test","type":"number","is_required":false}]`))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fields).To(gomega.HaveLen(1))
			})

			ginkgo.It("should handle nil input", func() {
				var fields persistenceInternal.Fields
				err := fields.Scan(nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(fields).To(gomega.BeEmpty())
			})

			ginkgo.It("should return error for invalid input", func() {
				var fields persistenceInternal.Fields
				err := fields.Scan(123)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("invalid type for fields"))
			})
		})
	})

	ginkgo.Context("ToDomain", func() {
		var internalActivity persistenceInternal.MaintenanceActivity

		ginkgo.BeforeEach(func() {
			now := utils.Time{Time: time.Now()}
			internalActivity = persistenceInternal.MaintenanceActivity{
				ID:                     "activity-id",
				Version:                1,
				TenantID:               "tenant-id",
				TypeName:               maintenanceDomain.ActivityTypeWaterSystem,
				CustomTypeName:         nil,
				Name:                   "Test Activity",
				Description:            "Test Description",
				Schedule:               "0 0 1 * *",
				NotificationDaysBefore: persistenceInternal.Days{7, 3},
				Fields:                 persistenceInternal.Fields{},
				IsActive:               true,
				CreatedAt:              now,
				UpdatedAt:              now,
				DeletedAt:              nil,
			}
		})

		ginkgo.When("converting to domain", func() {
			ginkgo.It("should convert basic fields correctly", func() {
				domain := internalActivity.ToDomain()
				gomega.Expect(domain.ID).To(gomega.Equal(shareddomain.ID("activity-id")))
				gomega.Expect(domain.Version).To(gomega.Equal(shareddomain.Version(1)))
				gomega.Expect(domain.TenantID).To(gomega.Equal(shareddomain.ID("tenant-id")))
				gomega.Expect(domain.Name).To(gomega.Equal(shareddomain.Name("Test Activity")))
				gomega.Expect(domain.Description).To(gomega.Equal(shareddomain.Description("Test Description")))
				gomega.Expect(domain.Schedule).To(gomega.Equal(maintenanceDomain.Schedule("0 0 1 * *")))
				gomega.Expect(domain.NotificationDaysBefore).To(gomega.Equal(maintenanceDomain.Days{7, 3}))
				gomega.Expect(domain.IsActive).To(gomega.BeTrue())
			})

			ginkgo.It("should handle CustomTypeName", func() {
				customType := "Custom Type"
				internalActivity.CustomTypeName = &customType
				domain := internalActivity.ToDomain()
				gomega.Expect(domain.CustomTypeName).NotTo(gomega.BeNil())
				gomega.Expect(string(*domain.CustomTypeName)).To(gomega.Equal("Custom Type"))
			})

			ginkgo.It("should handle DeletedAt", func() {
				now := utils.Time{Time: time.Now()}
				internalActivity.DeletedAt = &now
				domain := internalActivity.ToDomain()
				gomega.Expect(domain.DeletedAt).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("FromMaintenanceActivity", func() {
		var domainActivity maintenanceDomain.MaintenanceActivity

		ginkgo.BeforeEach(func() {
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields: []maintenanceDomain.FieldDefinition{
					{Name: shareddomain.Name("maintenance_type"), DisplayName: shareddomain.DisplayName("Maintenance Type"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
				},
			}

			domainActivity, _ = maintenanceDomain.NewMaintenanceActivityBuilder().
				WithTenantID(shareddomain.ID("tenant-id")).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithNotificationDaysBefore([]int{7, 3}).
				WithFields([]maintenanceDomain.FieldDefinition{
					{Name: shareddomain.Name("maintenance_type"), DisplayName: shareddomain.DisplayName("Maintenance Type"), Type: maintenanceDomain.FieldTypeText, IsRequired: true},
				}).
				Build()
		})

		ginkgo.When("converting from domain", func() {
			ginkgo.It("should convert basic fields correctly", func() {
				internal := persistenceInternal.FromMaintenanceActivity(domainActivity)
				gomega.Expect(internal.ID).To(gomega.Equal(domainActivity.ID.String()))
				gomega.Expect(internal.Version).To(gomega.Equal(int(domainActivity.Version)))
				gomega.Expect(internal.TenantID).To(gomega.Equal(domainActivity.TenantID.String()))
				gomega.Expect(internal.TypeName).To(gomega.Equal(string(domainActivity.Type.Name)))
				gomega.Expect(internal.Name).To(gomega.Equal(string(domainActivity.Name)))
				gomega.Expect(internal.Description).To(gomega.Equal(string(domainActivity.Description)))
				gomega.Expect(internal.Schedule).To(gomega.Equal(string(domainActivity.Schedule)))
				gomega.Expect(internal.NotificationDaysBefore).To(gomega.Equal(persistenceInternal.Days{7, 3}))
				gomega.Expect(internal.Fields).To(gomega.HaveLen(1))
				gomega.Expect(internal.IsActive).To(gomega.Equal(domainActivity.IsActive))
			})

			ginkgo.It("should handle CustomTypeName", func() {
				customTypeName := maintenanceDomain.CustomTypeName("Custom Type")
				domainActivity.CustomTypeName = &customTypeName
				internal := persistenceInternal.FromMaintenanceActivity(domainActivity)
				gomega.Expect(internal.CustomTypeName).NotTo(gomega.BeNil())
				gomega.Expect(*internal.CustomTypeName).To(gomega.Equal("Custom Type"))
			})

			ginkgo.It("should handle DeletedAt", func() {
				now := utils.Time{Time: time.Now()}
				domainActivity.DeletedAt = &now
				internal := persistenceInternal.FromMaintenanceActivity(domainActivity)
				gomega.Expect(internal.DeletedAt).NotTo(gomega.BeNil())
			})
		})
	})
})


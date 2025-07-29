package handlers_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication/handlers"
	"zensor-server/internal/infra/utils"
	mocksql "zensor-server/test/unit/doubles/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("ScheduledTaskHandler", func() {
	ginkgo.Context("TopicName", func() {
		var (
			handler *handlers.ScheduledTaskHandler
		)

		ginkgo.BeforeEach(func() {
			handler = handlers.NewScheduledTaskHandler(nil)
		})

		ginkgo.It("should return scheduled_tasks topic", func() {
			topic := handler.TopicName()
			gomega.Expect(topic).To(gomega.Equal(pubsub.Topic("scheduled_tasks")))
		})
	})

	ginkgo.Context("Create", func() {
		var (
			ctrl     *gomock.Controller
			mockOrm  *mocksql.MockORM
			handler  *handlers.ScheduledTaskHandler
			testData struct {
				ID               string
				Version          int
				TenantID         string
				DeviceID         string
				CommandTemplates string
				Schedule         string
				IsActive         bool
				CreatedAt        utils.Time
				UpdatedAt        utils.Time
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewScheduledTaskHandler(mockOrm)
			testData = struct {
				ID               string
				Version          int
				TenantID         string
				DeviceID         string
				CommandTemplates string
				Schedule         string
				IsActive         bool
				CreatedAt        utils.Time
				UpdatedAt        utils.Time
			}{
				ID:               "test-id",
				Version:          1,
				TenantID:         "tenant-1",
				DeviceID:         "device-1",
				CommandTemplates: `[{"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
				Schedule:         "* * * * *",
				IsActive:         true,
				CreatedAt:        utils.Time{Time: time.Now()},
				UpdatedAt:        utils.Time{Time: time.Now()},
			}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should create scheduled task successfully", func() {
			// Set up mock expectations
			mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
			mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
			mockOrm.EXPECT().Error().Return(nil)

			// Execute the method
			ctx := context.Background()
			err := handler.Create(ctx, "test-id", testData)

			// Assertions
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.ScheduledTaskHandler
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewScheduledTaskHandler(mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should get scheduled task by ID successfully", func() {
			// Set up mock expectations
			mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
			mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).DoAndReturn(
				func(dest interface{}, conds ...interface{}) *mocksql.MockORM {
					// Set the destination with test data
					scheduledTaskData := dest.(*handlers.ScheduledTaskData)
					*scheduledTaskData = handlers.ScheduledTaskData{
						ID:               "test-id",
						Version:          1,
						TenantID:         "tenant-1",
						DeviceID:         "device-1",
						CommandTemplates: `[{"port":15,"priority":"NORMAL","payload":{"index":1,"value":100},"wait_for":"5s"}]`,
						Schedule:         "* * * * *",
						IsActive:         true,
						CreatedAt:        utils.Time{Time: time.Now()},
						UpdatedAt:        utils.Time{Time: time.Now()},
					}
					return mockOrm
				},
			)
			mockOrm.EXPECT().Error().Return(nil)

			// Execute the method
			ctx := context.Background()
			result, err := handler.GetByID(ctx, "test-id")

			// Assertions
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result).NotTo(gomega.BeNil())

			resultMap, ok := result.(map[string]interface{})
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(resultMap["id"]).To(gomega.Equal("test-id"))
			gomega.Expect(resultMap["tenant_id"]).To(gomega.Equal("tenant-1"))
		})
	})

	ginkgo.Context("Update", func() {
		var (
			ctrl     *gomock.Controller
			mockOrm  *mocksql.MockORM
			handler  *handlers.ScheduledTaskHandler
			testData struct {
				ID               string
				Version          int
				TenantID         string
				DeviceID         string
				CommandTemplates string
				Schedule         string
				IsActive         bool
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewScheduledTaskHandler(mockOrm)
			testData = struct {
				ID               string
				Version          int
				TenantID         string
				DeviceID         string
				CommandTemplates string
				Schedule         string
				IsActive         bool
			}{
				ID:               "test-id",
				Version:          2,
				TenantID:         "tenant-1",
				DeviceID:         "device-1",
				CommandTemplates: `[{"port":15,"priority":"HIGH","payload":{"index":2,"value":200},"wait_for":"10s"}]`,
				Schedule:         "*/5 * * * *",
				IsActive:         false,
			}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should update scheduled task successfully", func() {
			// Set up mock expectations - WithContext is called twice in Update method
			mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm).Times(2)
			mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
			mockOrm.EXPECT().Error().Return(nil)
			mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
			mockOrm.EXPECT().Error().Return(nil)

			// Execute the method
			ctx := context.Background()
			err := handler.Update(ctx, "test-id", testData)

			// Assertions
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})

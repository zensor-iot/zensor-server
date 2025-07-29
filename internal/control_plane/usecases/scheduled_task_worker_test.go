package usecases_test

import (
	"context"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
	mockusecases "zensor-server/test/unit/doubles/control_plane/usecases"
	mockasync "zensor-server/test/unit/doubles/infra/async"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("ScheduledTaskWorker", func() {
	ginkgo.Context("NewScheduledTaskWorker", func() {
		var (
			ctrl                  *gomock.Controller
			mockScheduledTaskRepo *mockusecases.MockScheduledTaskRepository
			mockTaskService       *mockusecases.MockTaskService
			mockDeviceService     *mockusecases.MockDeviceService
			mockBroker            *mockasync.MockInternalBroker
			ticker                *time.Ticker
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockScheduledTaskRepo = mockusecases.NewMockScheduledTaskRepository(ctrl)
			mockTaskService = mockusecases.NewMockTaskService(ctrl)
			mockDeviceService = mockusecases.NewMockDeviceService(ctrl)
			mockBroker = mockasync.NewMockInternalBroker(ctrl)
			ticker = time.NewTicker(100 * time.Millisecond)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
			ticker.Stop()
		})

		ginkgo.It("should create a new scheduled task worker with mocks", func() {
			// Create a scheduled task worker
			worker := usecases.NewScheduledTaskWorker(
				ticker,
				mockScheduledTaskRepo,
				mockTaskService,
				mockDeviceService,
				nil, // TenantConfigurationService not available in mocks yet
				mockBroker,
			)

			// Verify the worker was created
			gomega.Expect(worker).NotTo(gomega.BeNil())
		})

		ginkgo.It("should verify mock repository interface", func() {
			// Test that the mock repository can be used
			testTenant := domain.Tenant{
				ID:   domain.ID("test-tenant"),
				Name: "Test Tenant",
			}
			testDevice := domain.Device{
				ID:   domain.ID("test-device"),
				Name: "Test Device",
			}
			testScheduledTask := domain.ScheduledTask{
				ID:               domain.ID("test-scheduled-task-123"),
				Version:          1,
				Tenant:           testTenant,
				Device:           testDevice,
				CommandTemplates: []domain.CommandTemplate{},
				Schedule:         "0 0 * * *",
				IsActive:         true,
				CreatedAt:        utils.Time{Time: time.Now()},
				UpdatedAt:        utils.Time{Time: time.Now()},
			}

			// Set up mock expectations
			mockScheduledTaskRepo.EXPECT().GetByID(gomock.Any(), domain.ID("test-scheduled-task-123")).Return(testScheduledTask, nil)
			mockScheduledTaskRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

			// Create a scheduled task worker
			worker := usecases.NewScheduledTaskWorker(
				ticker,
				mockScheduledTaskRepo,
				mockTaskService,
				mockDeviceService,
				nil, // TenantConfigurationService not available in mocks yet
				mockBroker,
			)
			gomega.Expect(worker).NotTo(gomega.BeNil())

			// Verify the mocks work correctly
			ctx := context.Background()
			task, err := mockScheduledTaskRepo.GetByID(ctx, domain.ID("test-scheduled-task-123"))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(task.ID).To(gomega.Equal(domain.ID("test-scheduled-task-123")))

			err = mockScheduledTaskRepo.Update(ctx, testScheduledTask)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})

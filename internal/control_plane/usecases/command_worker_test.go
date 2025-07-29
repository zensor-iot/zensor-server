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

var _ = ginkgo.Describe("CommandWorker", func() {
	ginkgo.Context("NewCommandWorker", func() {
		var (
			ctrl       *gomock.Controller
			mockRepo   *mockusecases.MockCommandRepository
			mockBroker *mockasync.MockInternalBroker
			ticker     *time.Ticker
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockRepo = mockusecases.NewMockCommandRepository(ctrl)
			mockBroker = mockasync.NewMockInternalBroker(ctrl)
			ticker = time.NewTicker(100 * time.Millisecond)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
			ticker.Stop()
		})

		ginkgo.It("should create a new command worker with mocks", func() {
			// Create a command worker
			worker := usecases.NewCommandWorker(ticker, mockRepo, mockBroker)

			// Verify the worker was created
			gomega.Expect(worker).NotTo(gomega.BeNil())
		})

		ginkgo.It("should verify mock repository interface", func() {
			// Test that the mock repository can be used
			testCommand := domain.Command{
				ID:       "test-command-123",
				Version:  1,
				Device:   domain.Device{ID: "device-1", Name: "Test Device"},
				Task:     domain.Task{ID: "task-1"},
				Port:     15,
				Priority: domain.CommandPriority("NORMAL"),
				Payload: domain.CommandPayload{
					Index: 1,
					Value: 100,
				},
				DispatchAfter: utils.Time{Time: time.Now()},
				CreatedAt:     utils.Time{Time: time.Now()},
				Ready:         true,
				Sent:          false,
				Status:        domain.CommandStatusPending,
			}

			// Set up mock expectations
			mockRepo.EXPECT().GetByID(gomock.Any(), domain.ID("test-command-123")).Return(testCommand, nil)
			mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

			// Create a command worker
			worker := usecases.NewCommandWorker(ticker, mockRepo, mockBroker)
			gomega.Expect(worker).NotTo(gomega.BeNil())

			// Verify the mocks work correctly
			ctx := context.Background()
			cmd, err := mockRepo.GetByID(ctx, domain.ID("test-command-123"))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(cmd.ID).To(gomega.Equal(domain.ID("test-command-123")))

			err = mockRepo.Update(ctx, testCommand)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})

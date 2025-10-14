package replication_test

import (
	"context"
	"log/slog"
	"os"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication"
	mockpubsub "zensor-server/test/unit/doubles/infra/pubsub"
	mocksql "zensor-server/test/unit/doubles/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("Service", func() {
	ginkgo.BeforeEach(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(100)})))
	})
	ginkgo.Context("NewService", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should create a new service", func() {
			service := replication.NewService(mockConsumerFactory, mockOrm)

			gomega.Expect(service).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Start", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
			service             *replication.Service
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			service = replication.NewService(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			service.Stop()
			ctrl.Finish()
		})

		ginkgo.When("starting with no handlers", func() {
			ginkgo.It("should start successfully", func() {
				// Start with no handlers (should succeed)
				err := service.Start()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("starting with handlers", func() {
			ginkgo.It("should start with handlers successfully", func() {
				// Register a handler
				handler := &ServiceMockTopicHandler{}

				// Set up mock expectations to prevent goroutine failures
				consumer := mockpubsub.NewMockConsumer(ctrl)
				consumer.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockConsumerFactory.EXPECT().New().Return(consumer).AnyTimes()

				err := service.RegisterHandler(handler)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Start the service
				err = service.Start()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("Stop", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
			service             *replication.Service
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			service = replication.NewService(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			service.Stop()
			ctrl.Finish()
		})

		ginkgo.It("should stop service successfully", func() {
			// Start the service
			err := service.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Stop the service
			service.Stop()

			// No assertions needed - just checking it doesn't panic
		})
	})

	ginkgo.Context("RegisterHandler", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
			service             *replication.Service
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			service = replication.NewService(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			service.Stop()
			ctrl.Finish()
		})

		ginkgo.It("should register handler successfully", func() {
			// Create a mock handler
			handler := &ServiceMockTopicHandler{}

			// Register the handler
			err := service.RegisterHandler(handler)

			// Assertions
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})

// ServiceMockTopicHandler is a simple mock implementation for testing
type ServiceMockTopicHandler struct{}

func (m *ServiceMockTopicHandler) TopicName() pubsub.Topic {
	return "test-topic"
}

func (m *ServiceMockTopicHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	return nil
}

func (m *ServiceMockTopicHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	return nil, nil
}

func (m *ServiceMockTopicHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	return nil
}

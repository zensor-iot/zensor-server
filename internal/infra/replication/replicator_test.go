package replication_test

import (
	"context"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication"
	mockpubsub "zensor-server/test/unit/doubles/infra/pubsub"
	mocksql "zensor-server/test/unit/doubles/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("Replicator", func() {
	ginkgo.Context("NewReplicator", func() {
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

		ginkgo.It("should create a new replicator", func() {
			replicator := replication.NewReplicator(mockConsumerFactory, mockOrm)
			gomega.Expect(replicator).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("RegisterHandler", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
			replicator          *replication.Replicator
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			replicator = replication.NewReplicator(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("registering a new handler", func() {
			ginkgo.It("should register handler successfully", func() {
				// Create a mock handler
				mockHandler := &MockTopicHandler{}

				// Register the handler
				err := replicator.RegisterHandler(mockHandler)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("registering a duplicate handler", func() {
			ginkgo.It("should return error for duplicate handler", func() {
				// Create a mock handler
				mockHandler := &MockTopicHandler{}

				// Register the handler twice
				err1 := replicator.RegisterHandler(mockHandler)
				err2 := replicator.RegisterHandler(mockHandler)

				// Assertions
				gomega.Expect(err1).NotTo(gomega.HaveOccurred())
				gomega.Expect(err2).To(gomega.HaveOccurred())
				gomega.Expect(err2.Error()).To(gomega.ContainSubstring("handler already registered"))
			})
		})
	})

	ginkgo.Context("Start", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockConsumer        *mockpubsub.MockConsumer
			mockOrm             *mocksql.MockORM
			replicator          *replication.Replicator
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockConsumer = mockpubsub.NewMockConsumer(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			replicator = replication.NewReplicator(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("starting with no handlers", func() {
			ginkgo.It("should start without error", func() {
				// Start the replicator
				err := replicator.Start()

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("starting with handlers", func() {
			ginkgo.It("should start with handlers successfully", func() {
				// Create a mock handler
				mockHandler := &MockTopicHandler{}

				// Register the handler
				err := replicator.RegisterHandler(mockHandler)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Set up mock expectations
				mockConsumerFactory.EXPECT().New().Return(mockConsumer)
				mockConsumer.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				// Start the replicator
				err = replicator.Start()

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})

	ginkgo.Context("Stop", func() {
		var (
			ctrl                *gomock.Controller
			mockConsumerFactory *mockpubsub.MockConsumerFactory
			mockOrm             *mocksql.MockORM
			replicator          *replication.Replicator
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockConsumerFactory = mockpubsub.NewMockConsumerFactory(ctrl)
			mockOrm = mocksql.NewMockORM(ctrl)
			replicator = replication.NewReplicator(mockConsumerFactory, mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should stop replicator successfully", func() {
			// Start the replicator
			err := replicator.Start()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Stop the replicator
			replicator.Stop()

			// No assertions needed - just checking it doesn't panic
		})
	})
})

// MockTopicHandler is a simple mock implementation for testing
type MockTopicHandler struct{}

func (m *MockTopicHandler) TopicName() pubsub.Topic {
	return "test-topic"
}

func (m *MockTopicHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	return nil
}

func (m *MockTopicHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	return nil, nil
}

func (m *MockTopicHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	return nil
}

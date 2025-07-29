package communication_test

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"zensor-server/internal/control_plane/communication"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
	mockpubsub "zensor-server/test/unit/doubles/infra/pubsub"
)

var _ = ginkgo.Describe("CommandPublisher", func() {
	var (
		ctrl          *gomock.Controller
		mockFactory   *mockpubsub.MockPublisherFactory
		mockPublisher *mockpubsub.MockPublisher
		publisher     *communication.CommandPublisher
		testCommand   domain.Command
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockFactory = mockpubsub.NewMockPublisherFactory(ctrl)
		mockPublisher = mockpubsub.NewMockPublisher(ctrl)

		// Set up mock expectations for constructor
		mockFactory.EXPECT().New(pubsub.Topic("device_commands"), gomock.Any()).Return(mockPublisher, nil)

		// Create the command publisher
		var err error
		publisher, err = communication.NewCommandPublisher(mockFactory)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Create a test domain command
		testCommand = domain.Command{
			ID:       domain.ID("test-command-id"),
			Version:  domain.Version(1),
			Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
			Task:     domain.Task{ID: domain.ID("test-task-id")},
			Port:     domain.Port(1),
			Priority: domain.CommandPriority("NORMAL"),
			Payload: domain.CommandPayload{
				Index: domain.Index(0),
				Value: domain.CommandValue(100),
			},
			DispatchAfter: utils.Time{Time: time.Now()},
			Ready:         true,
			Sent:          false,
			SentAt:        utils.Time{},
		}
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("Dispatch", func() {
		ginkgo.When("dispatching a valid command", func() {
			ginkgo.It("should publish the command with correct Avro format", func() {
				// Set up mock expectations for the dispatch call
				mockPublisher.EXPECT().Publish(gomock.Any(), pubsub.Key(testCommand.ID), gomock.Any()).Return(nil)

				// Dispatch the command
				err := publisher.Dispatch(context.Background(), testCommand)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
	})
})

package communication_test

import (
	"context"
	"time"
	"zensor-server/internal/control_plane/communication"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("CommandPublisher", func() {
	var (
		mockFactory *mockPublisherFactory
		publisher   *communication.CommandPublisher
		testCommand domain.Command
	)

	ginkgo.BeforeEach(func() {
		// Create a mock publisher factory
		mockFactory = &mockPublisherFactory{
			publisher: &mockPublisher{},
		}

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

	ginkgo.Context("Dispatch", func() {
		ginkgo.When("dispatching a valid command", func() {
			ginkgo.It("should publish the command with correct Avro format", func() {
				// Dispatch the command
				err := publisher.Dispatch(context.Background(), testCommand)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify that the mock publisher was called with the correct data
				mockPub := mockFactory.publisher.(*mockPublisher)
				gomega.Expect(mockPub.publishedKey).To(gomega.Equal(pubsub.Key(testCommand.ID)))

				// Verify that the published value is an AvroCommand
				avroCmd, ok := mockPub.publishedValue.(*avro.AvroCommand)
				gomega.Expect(ok).To(gomega.BeTrue(), "Expected published value to be *avro.AvroCommand")

				// Verify the AvroCommand fields match the domain command
				gomega.Expect(avroCmd.ID).To(gomega.Equal(string(testCommand.ID)))
				gomega.Expect(avroCmd.Version).To(gomega.Equal(int(testCommand.Version) + 1)) // Version should be incremented
				gomega.Expect(avroCmd.DeviceID).To(gomega.Equal(string(testCommand.Device.ID)))
				gomega.Expect(avroCmd.DeviceName).To(gomega.Equal(testCommand.Device.Name))
				gomega.Expect(avroCmd.TaskID).To(gomega.Equal(string(testCommand.Task.ID)))
				gomega.Expect(avroCmd.PayloadIndex).To(gomega.Equal(int(testCommand.Payload.Index)))
				gomega.Expect(avroCmd.PayloadValue).To(gomega.Equal(int(testCommand.Payload.Value)))
				gomega.Expect(avroCmd.Port).To(gomega.Equal(int(testCommand.Port)))
				gomega.Expect(avroCmd.Priority).To(gomega.Equal(string(testCommand.Priority)))
				gomega.Expect(avroCmd.Ready).To(gomega.Equal(testCommand.Ready))
				gomega.Expect(avroCmd.Sent).To(gomega.Equal(testCommand.Sent))
			})
		})
	})
})

// Mock implementations for testing

type mockPublisherFactory struct {
	publisher pubsub.Publisher
}

func (f *mockPublisherFactory) New(topic pubsub.Topic, prototype pubsub.Message) (pubsub.Publisher, error) {
	return f.publisher, nil
}

type mockPublisher struct {
	publishedKey   pubsub.Key
	publishedValue pubsub.Message
}

func (p *mockPublisher) Publish(ctx context.Context, key pubsub.Key, value pubsub.Message) error {
	p.publishedKey = key
	p.publishedValue = value
	return nil
}

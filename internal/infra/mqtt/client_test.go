package mqtt_test

import (
	"context"
	"zensor-server/internal/infra/mqtt"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MQTT Client", func() {
	ginkgo.Context("SimpleClientOpts", func() {
		var opts mqtt.SimpleClientOpts

		ginkgo.When("creating client options", func() {
			ginkgo.BeforeEach(func() {
				opts = mqtt.SimpleClientOpts{
					Broker:   "tcp://localhost:1883",
					ClientID: "test-client",
					Username: "test-user",
					Password: "test-pass",
				}
			})

			ginkgo.It("should have correct configuration values", func() {
				gomega.Expect(opts.Broker).To(gomega.Equal("tcp://localhost:1883"))
				gomega.Expect(opts.ClientID).To(gomega.Equal("test-client"))
				gomega.Expect(opts.Username).To(gomega.Equal("test-user"))
				gomega.Expect(opts.Password).To(gomega.Equal("test-pass"))
			})
		})
	})

	ginkgo.Context("MessageTypeAlias", func() {
		ginkgo.When("checking message type alias", func() {
			ginkgo.It("should properly alias Message to paho.Message", func() {
				// This test ensures that Message is properly aliased to paho.Message
				var _ mqtt.Message = (paho.Message)(nil)
			})
		})
	})

	ginkgo.Context("NewNoOpClient", func() {
		var client mqtt.Client

		ginkgo.When("used as Client", func() {
			ginkgo.BeforeEach(func() {
				client = mqtt.NewNoOpClient()
			})

			ginkgo.It("should subscribe without error", func() {
				err := client.Subscribe("any/topic", 0, func(mqtt.Client, mqtt.Message) {})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should publish without error when context is active", func() {
				err := client.Publish(context.Background(), "any/topic", map[string]string{"k": "v"})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return context error when context is cancelled", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err := client.Publish(ctx, "any/topic", "payload")
				gomega.Expect(err).To(gomega.Equal(context.Canceled))
			})

			ginkgo.It("should disconnect without panic", func() {
				client.Disconnect()
			})
		})
	})
})

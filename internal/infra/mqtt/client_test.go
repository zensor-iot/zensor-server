package mqtt_test

import (
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
})

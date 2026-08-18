package usecases_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/victron/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type fakeMessage struct {
	topic   string
	payload []byte
}

func (m fakeMessage) Duplicate() bool   { return false }
func (m fakeMessage) Qos() byte         { return 0 }
func (m fakeMessage) Retained() bool    { return false }
func (m fakeMessage) Topic() string     { return m.topic }
func (m fakeMessage) MessageID() uint16 { return 0 }
func (m fakeMessage) Payload() []byte   { return m.payload }
func (m fakeMessage) Ack()              {}

type fakeMQTTClient struct {
	mu           sync.Mutex
	callback     mqtt.MessageHandler
	keepAlives   int
	publishError error
}

func (c *fakeMQTTClient) Subscribe(_ string, _ byte, callback mqtt.MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callback = callback
	return nil
}

func (c *fakeMQTTClient) Publish(_ context.Context, topic string, _ any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.HasSuffix(topic, "/keepalive") {
		c.keepAlives++
	}
	return c.publishError
}

func (c *fakeMQTTClient) Disconnect() {}

func (c *fakeMQTTClient) keepAliveCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.keepAlives
}

func (c *fakeMQTTClient) failPublishes(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.publishError = err
}

func (c *fakeMQTTClient) deliver(topic string, payload string) {
	c.mu.Lock()
	callback := c.callback
	c.mu.Unlock()

	gomega.Expect(callback).NotTo(gomega.BeNil())
	callback(c, fakeMessage{topic: topic, payload: []byte(payload)})
}

var _ mqtt.Client = (*fakeMQTTClient)(nil)

var _ = ginkgo.Describe("VictronWorker", func() {
	var (
		client *fakeMQTTClient
		broker *async.LocalBroker
		worker *usecases.VictronWorker
		cancel context.CancelFunc
	)

	ginkgo.BeforeEach(func() {
		client = &fakeMQTTClient{}
		broker = async.NewLocalBroker()
		worker = usecases.NewVictronWorker("d41243b4e8e4", client, broker)

		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		go worker.Run(ctx, func() {})
	})

	ginkgo.AfterEach(func() {
		cancel()
	})

	ginkgo.Context("requesting the initial full publish", func() {
		ginkgo.When("the GX has not confirmed a full publish", func() {
			ginkgo.It("should retry well before the regular keep-alive interval", func() {
				gomega.Eventually(client.keepAliveCount, 5*time.Second, 100*time.Millisecond).
					Should(gomega.BeNumerically(">=", 3))
			})
		})

		ginkgo.When("the broker rejects the keep-alive while it is still connecting", func() {
			ginkgo.BeforeEach(func() {
				client.failPublishes(errors.New("not connected"))
			})

			ginkgo.It("should keep retrying instead of giving up", func() {
				gomega.Eventually(client.keepAliveCount, 5*time.Second, 100*time.Millisecond).
					Should(gomega.BeNumerically(">=", 3))
			})
		})

		ginkgo.When("the GX confirms the full publish", func() {
			ginkgo.It("should settle back to the regular keep-alive interval", func() {
				gomega.Eventually(client.keepAliveCount, 5*time.Second, 100*time.Millisecond).
					Should(gomega.BeNumerically(">=", 2))

				client.deliver("N/d41243b4e8e4/full_publish_completed", "")

				gomega.Eventually(func() bool {
					before := client.keepAliveCount()
					time.Sleep(3 * time.Second)
					return client.keepAliveCount() == before
				}, 15*time.Second, time.Second).Should(gomega.BeTrue())
			})
		})
	})

	ginkgo.Context("non telemetry topics", func() {
		ginkgo.When("the GX publishes its heartbeat", func() {
			ginkgo.It("should not treat it as telemetry", func() {
				gomega.Eventually(client.keepAliveCount, 5*time.Second, 100*time.Millisecond).
					Should(gomega.BeNumerically(">=", 1))

				gomega.Expect(func() {
					client.deliver("N/d41243b4e8e4/heartbeat", "")
				}).NotTo(gomega.Panic())
			})
		})
	})
})

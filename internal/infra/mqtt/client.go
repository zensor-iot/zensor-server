package mqtt

import (
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client interface {
	Subscribe(topic string, qos byte, callback MessageHandler) error
	Disconnect()
}

var _ Client = &SimpleClient{}

func NewSimpleClient() *SimpleClient {
	opts := paho.NewClientOptions().
		AddBroker("nam1.cloud.thethings.network:1883").
		SetClientID("local_client").
		SetUsername("my-new-application-2021@ttn").
		SetPassword("NNSXS.CJWYCNW4FRNQKDDCXM27ZPKMQ6UBFR2RWBHFSUA.V7O6N6WUUIBXJUZ7LZCRW5SISWUIP4KIZOZHA2OJ4QBQVFSPAJQQ").
		SetOnConnectHandler(onConnectHandler)

	client := paho.NewClient(opts)
	token := client.Connect()
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		panic(token.Error())
	}

	return &SimpleClient{
		client,
	}
}

type SimpleClient struct {
	client paho.Client
}

func onConnectHandler(_ paho.Client) {
	slog.Info("connected to MQTT broker")
}

func (c *SimpleClient) Subscribe(topic string, qos byte, callback MessageHandler) error {
	pahoCallback := func(_ paho.Client, msg paho.Message) {
		callback(c, msg)
	}
	token := c.client.Subscribe(topic, qos, pahoCallback)
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		return fmt.Errorf("subscribing to topic %s: %w", topic, token.Error())
	}

	return nil
}

type MessageHandler func(Client, Message)

type Message interface {
	Topic() string
	MessageID() uint16
	Payload() []byte
	Ack()
}

func (c *SimpleClient) Disconnect() {
	waitForInMilliseconds := 5 * 1000
	c.client.Disconnect(uint(waitForInMilliseconds))
}

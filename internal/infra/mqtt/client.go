package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	_defaultQoS      = 0 // At most once
	_defaultRetained = false
	_publishTimeout  = 5 * time.Second
)

type Client interface {
	Subscribe(topic string, qos byte, callback MessageHandler) error
	Publish(topic string, msg any) error

	Disconnect()
}

type SimpleClientOpts struct {
	Broker   string
	ClientID string
	Username string
	Password string
}

func NewSimpleClient(opts SimpleClientOpts) *SimpleClient {
	pahoOpts := paho.NewClientOptions().
		AddBroker(opts.Broker).
		SetClientID(opts.ClientID).
		SetUsername(opts.Username).
		SetPassword(opts.Password).
		SetOnConnectHandler(onConnectHandler)

	client := paho.NewClient(pahoOpts)
	token := client.Connect()
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		panic(token.Error())
	}

	return &SimpleClient{
		client,
	}
}

var _ Client = (*SimpleClient)(nil)

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

func (c *SimpleClient) Publish(topic string, msg any) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}
	token := c.client.Publish(topic, _defaultQoS, _defaultRetained, payload)
	token.WaitTimeout(_publishTimeout)
	if token.Error() != nil {
		return fmt.Errorf("publishing to topic %s: %w", topic, token.Error())
	}

	return nil
}

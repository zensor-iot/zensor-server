package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
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

// Subscription tracks a topic subscription for reconnection recovery
type subscription struct {
	topic    string
	qos      byte
	callback MessageHandler
}

func NewSimpleClient(opts SimpleClientOpts) *SimpleClient {
	simpleClient := &SimpleClient{
		subscriptions: make(map[string]subscription),
		mu:            sync.RWMutex{},
	}

	onConnectHandler := func(client paho.Client) {
		slog.Info("connected to MQTT broker")
		simpleClient.resubscribeAll(client)
	}

	onConnectionLostHandler := func(_ paho.Client, err error) {
		slog.Error("connection lost to MQTT broker", "error", err)
	}

	pahoOpts := paho.NewClientOptions().
		AddBroker(opts.Broker).
		SetClientID(opts.ClientID).
		SetUsername(opts.Username).
		SetPassword(opts.Password).
		SetOnConnectHandler(onConnectHandler).
		SetAutoReconnect(true).
		SetConnectionLostHandler(onConnectionLostHandler).
		SetKeepAlive(10 * time.Second).
		SetConnectTimeout(5 * time.Second)

	client := paho.NewClient(pahoOpts)
	token := client.Connect()
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		panic(token.Error())
	}

	simpleClient.client = client
	return simpleClient
}

var _ Client = (*SimpleClient)(nil)

type SimpleClient struct {
	client        paho.Client
	subscriptions map[string]subscription
	mu            sync.RWMutex
}

// resubscribeAll re-establishes all subscriptions after reconnection
func (c *SimpleClient) resubscribeAll(client paho.Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.subscriptions) == 0 {
		slog.Debug("no subscriptions to restore")
		return
	}

	slog.Info("restoring MQTT subscriptions after reconnection", "count", len(c.subscriptions))

	for topic, sub := range c.subscriptions {
		pahoCallback := func(_ paho.Client, msg paho.Message) {
			sub.callback(c, msg)
		}

		token := client.Subscribe(sub.topic, sub.qos, pahoCallback)
		token.WaitTimeout(5 * time.Second)
		if token.Error() != nil {
			slog.Error("failed to restore subscription after reconnection",
				"topic", topic, "error", token.Error())
		} else {
			slog.Debug("subscription restored", "topic", topic)
		}
	}
}

func (c *SimpleClient) Subscribe(topic string, qos byte, callback MessageHandler) error {
	// Store subscription for reconnection recovery
	c.mu.Lock()
	c.subscriptions[topic] = subscription{
		topic:    topic,
		qos:      qos,
		callback: callback,
	}
	c.mu.Unlock()

	pahoCallback := func(_ paho.Client, msg paho.Message) {
		callback(c, msg)
	}
	token := c.client.Subscribe(topic, qos, pahoCallback)
	token.WaitTimeout(5 * time.Second)
	if token.Error() != nil {
		// Remove from subscriptions if subscribe failed
		c.mu.Lock()
		delete(c.subscriptions, topic)
		c.mu.Unlock()
		return fmt.Errorf("subscribing to topic %s: %w", topic, token.Error())
	}

	slog.Info("subscribed to MQTT topic", "topic", topic, "qos", qos)
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
	// Clear subscriptions on manual disconnect
	c.mu.Lock()
	c.subscriptions = make(map[string]subscription)
	c.mu.Unlock()

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

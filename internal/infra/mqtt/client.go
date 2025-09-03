package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	_defaultQoS           = 0 // At most once
	_defaultRetained      = false
	_publishTimeout       = 5 * time.Second
	_maxReconnectInterval = 1 * time.Minute
)

type Client interface {
	Subscribe(topic string, qos byte, callback MessageHandler) error
	Publish(ctx context.Context, topic string, msg any) error

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

	// Generate a unique client ID to prevent session conflicts
	uniqueClientID := fmt.Sprintf("%s-%s", opts.ClientID, uuid.NewString()[:8])

	onConnectHandler := func(client paho.Client) {
		slog.Info("connected to MQTT broker", "client_id", uniqueClientID)
		simpleClient.resubscribeAll(client)
	}

	onConnectionLostHandler := func(_ paho.Client, err error) {
		slog.Error("connection lost to MQTT broker", "error", err, "client_id", uniqueClientID)
	}

	pahoOpts := paho.NewClientOptions().
		AddBroker(opts.Broker).
		SetClientID(uniqueClientID).
		SetUsername(opts.Username).
		SetPassword(opts.Password).
		SetOnConnectHandler(onConnectHandler).
		SetAutoReconnect(true).
		SetConnectionLostHandler(onConnectionLostHandler).
		SetKeepAlive(10 * time.Second).
		SetConnectTimeout(5 * time.Second).
		SetCleanSession(true).                          // Always start with a clean session
		SetMaxReconnectInterval(_maxReconnectInterval). // Limit reconnection attempts
		SetDefaultPublishHandler(func(client paho.Client, msg paho.Message) {
			slog.Debug("received message on default handler", "topic", msg.Topic())
		})

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
	processedMsgs sync.Map // Track processed message IDs to prevent duplicates
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
			// Check for duplicate messages
			msgKey := fmt.Sprintf("%s-%d", msg.Topic(), msg.MessageID())
			if _, exists := c.processedMsgs.LoadOrStore(msgKey, true); exists {
				slog.Debug("duplicate message ignored", "topic", msg.Topic(), "message_id", msg.MessageID())
				msg.Ack() // Still acknowledge to prevent retransmission
				return
			}

			sub.callback(c, msg)
			msg.Ack()

			// Clean up processed message tracking after some time
			go func(key string) {
				time.Sleep(30 * time.Second)
				c.processedMsgs.Delete(key)
			}(msgKey)
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
		// Check for duplicate messages
		msgKey := fmt.Sprintf("%s-%d", msg.Topic(), msg.MessageID())
		if _, exists := c.processedMsgs.LoadOrStore(msgKey, true); exists {
			slog.Debug("duplicate message ignored", "topic", msg.Topic(), "message_id", msg.MessageID())
			msg.Ack() // Still acknowledge to prevent retransmission
			return
		}

		callback(c, msg)
		msg.Ack()

		// Clean up processed message tracking after some time
		go func(key string) {
			time.Sleep(30 * time.Second)
			c.processedMsgs.Delete(key)
		}(msgKey)
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

type Message = paho.Message

func (c *SimpleClient) Disconnect() {
	// Clear subscriptions on manual disconnect
	c.mu.Lock()
	c.subscriptions = make(map[string]subscription)
	c.mu.Unlock()

	// Clear processed messages tracking
	c.processedMsgs = sync.Map{}

	waitForInMilliseconds := 5 * 1000
	c.client.Disconnect(uint(waitForInMilliseconds))
}

func (c *SimpleClient) Publish(ctx context.Context, topic string, msg any) error {
	tracer := otel.Tracer("zensor-server")
	ctx, span := tracer.Start(ctx, "mqtt.publish",
		trace.WithAttributes(
			attribute.String("span.kind", "client"),
			attribute.String("component", "mqtt-client"),
			attribute.String("messaging.system", "mqtt"),
			attribute.String("messaging.destination", topic),
			attribute.String("messaging.destination_kind", "topic"),
		),
	)
	defer span.End()

	payload, err := json.Marshal(msg)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("marshaling message: %w", err)
	}

	token := c.client.Publish(topic, _defaultQoS, _defaultRetained, payload)
	token.WaitTimeout(_publishTimeout)
	if token.Error() != nil {
		span.RecordError(token.Error())
		return fmt.Errorf("publishing to topic %s: %w", topic, token.Error())
	}

	return nil
}

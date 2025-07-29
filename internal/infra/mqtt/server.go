package mqtt

import (
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	username   string = "zensor_server"
	qos_level  byte   = 0
	maxRetries int    = 10

	EventTypeMessage  string = "message"
	EventTypePresence string = "presence"
	EventTypeMyself   string = "myself"
)

type Event struct {
	DeviceID string
	Type     string
	Value    []byte
}

type mqttPeer struct {
	id              string
	client          mqtt.Client
	topic           string
	outboundChannel chan Event
}

func Run(broker string, topic string, outboundChannel chan Event) {
	id := uuid.NewString()
	peer := &mqttPeer{
		id,
		connect(broker, id),
		topic,
		outboundChannel,
	}

	peer.outboundChannel <- Event{id, EventTypeMyself, nil}

	peer.subscribe()
}

func connect(broker, id string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetUsername(username)
	for try := 0; try < maxRetries; try++ {
		client := mqtt.NewClient(opts)
		token := client.Connect()

		if token.Wait() && token.Error() != nil {
			slog.Info("error connecting to mqtt broker: %s", slog.String("error", token.Error().Error()))
			slog.Info("retrying... ")
			time.Sleep(5 * time.Second)
		} else {
			return client
		}
	}

	slog.Info("imposible to connect to mqtt broker after many retries", slog.Int("retries", maxRetries))
	return nil
}

func (p *mqttPeer) subscribe() {
	slog.Info("subscribing to channel", "topic", p.topic)
	p.client.Subscribe(p.topic, qos_level, p.onMessageReceive)
}

func (p *mqttPeer) onMessageReceive(c mqtt.Client, msg mqtt.Message) {
	slog.Info("message received", "msg", msg.Payload())
	p.outboundChannel <- Event{p.id, EventTypeMessage, msg.Payload()}
	msg.Ack()
}

func (p *mqttPeer) publish(_ string) {
	slog.Info("not implemented yet")
}

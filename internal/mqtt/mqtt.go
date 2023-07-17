package mqtt

import (
	"time"

	"zensor-server/internal/logger"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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
	client          MQTT.Client
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

func connect(broker, id string) MQTT.Client {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetUsername(username)
	for try := 0; try < maxRetries; try++ {
		client := MQTT.NewClient(opts)
		token := client.Connect()

		if token.Wait() && token.Error() != nil {
			logger.Info("error connecting to mqtt broker: %s", token.Error())
			logger.Info("retrying... ")
			time.Sleep(5 * time.Second)
		} else {
			return client
		}
	}

	logger.Info("imposible to connect to mqtt broker after %d retries", maxRetries)
	return nil
}

func (p *mqttPeer) subscribe() {
	logger.Info("subscribing to channel", "topic", p.topic)
	p.client.Subscribe(p.topic, qos_level, p.onMessageReceive)
}

func (p *mqttPeer) onMessageReceive(c MQTT.Client, msg MQTT.Message) {
	logger.Info("message received", "msg", msg.Payload())
	p.outboundChannel <- Event{p.id, EventTypeMessage, msg.Payload()}
}

func (p *mqttPeer) publish(message string) {
	logger.Info("not implemented yet")
}

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"zensor-server/internal/infra/mqtt"
)

var topics = []string{
	"join",
	"up",
	"down/queued",
	"down/sent",
	"down/failed",
	"down/ack",
}

func main() {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})),
	)
	slog.Info("application starting")

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	simpleClientOpts := mqtt.SimpleClientOpts{
		Broker:   "dummy",
		ClientID: "dummy",
		Username: "dummy",
		Password: "dummy", //pragma: allowlist secret
	}
	mqttClient := mqtt.NewSimpleClient(simpleClientOpts)
	var (
		topicBase           = "v3/my-new-application-2021@ttn/devices/wireless-stick-seba"
		qos            byte = 0
		messageHandler      = func(_ mqtt.Client, msg mqtt.Message) {
			slog.Info("message received",
				slog.String("topic", msg.Topic()),
				slog.Uint64("message_id", uint64(msg.MessageID())),
				slog.String("payload", string(msg.Payload())),
			)
		}
	)

	for _, suffix := range topics {
		topic := fmt.Sprintf("%s/%s", topicBase, suffix)
		slog.Debug("final topic", slog.String("value", topic))
		mqttClient.Subscribe(topic, qos, messageHandler)
	}

	<-signalChannel
	mqttClient.Disconnect()
	slog.Info("good bye!!!")
	os.Exit(0)
}

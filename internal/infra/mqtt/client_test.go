package mqtt

import (
	"testing"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestSimpleClientOpts(t *testing.T) {
	opts := SimpleClientOpts{
		Broker:   "tcp://localhost:1883",
		ClientID: "test-client",
		Username: "test-user",
		Password: "test-pass",
	}

	assert.Equal(t, "tcp://localhost:1883", opts.Broker)
	assert.Equal(t, "test-client", opts.ClientID)
	assert.Equal(t, "test-user", opts.Username)
	assert.Equal(t, "test-pass", opts.Password)
}

func TestMessageTypeAlias(t *testing.T) {
	// This test ensures that Message is properly aliased to paho.Message
	var _ Message = (paho.Message)(nil)
}

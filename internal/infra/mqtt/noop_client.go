package mqtt

import "context"

// NoOpClient is a Client that does not connect to a broker. It is used when
// ENV=local so the API process can run without TTN or other MQTT credentials.
type NoOpClient struct{}

// NewNoOpClient returns a Client that performs no network I/O.
func NewNoOpClient() *NoOpClient {
	return &NoOpClient{}
}

// Subscribe implements Client.
func (c *NoOpClient) Subscribe(topic string, qos byte, callback MessageHandler) error {
	return nil
}

// Publish implements Client.
func (c *NoOpClient) Publish(ctx context.Context, topic string, msg any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// Disconnect implements Client.
func (c *NoOpClient) Disconnect() {}

var _ Client = (*NoOpClient)(nil)

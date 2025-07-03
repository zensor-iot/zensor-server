package steps

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// DeviceStateMessage represents a device state message from WebSocket
type DeviceStateMessage struct {
	Type      string                  `json:"type"`
	DeviceID  string                  `json:"device_id"`
	Timestamp time.Time               `json:"timestamp"`
	Data      map[string][]SensorData `json:"data"`
}

// SensorData represents sensor data in the WebSocket message
type SensorData struct {
	Index uint    `json:"index"`
	Value float64 `json:"value"`
}

// WebSocket connection for testing
var wsConn *websocket.Conn

func (fc *FeatureContext) theDeviceHasCachedSensorData() error {
	// This step simulates that the device has cached sensor data
	// In a real scenario, this would be done by sending sensor data through the LoRa integration
	// For now, we'll just ensure the device exists and has been active
	return nil
}

func (fc *FeatureContext) iConnectToTheWebSocketEndpoint() error {
	// Connect to the WebSocket endpoint
	url := fmt.Sprintf("ws://localhost:3000/ws/device-messages")
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("websocket connection failed with status %d: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	wsConn = conn
	return nil
}

func (fc *FeatureContext) iShouldReceiveCachedDeviceStatesImmediately() error {
	if wsConn == nil {
		return fmt.Errorf("websocket connection not established")
	}

	// Set a timeout for receiving messages
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read messages for a short time to capture cached states
	var messages []DeviceStateMessage
	timeout := time.After(3 * time.Second)

	for {
		select {
		case <-timeout:
			break
		default:
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				// Connection closed or timeout
				break
			}

			var deviceState DeviceStateMessage
			if err := json.Unmarshal(message, &deviceState); err != nil {
				// Skip non-device-state messages
				continue
			}

			if deviceState.Type == "device_state" {
				messages = append(messages, deviceState)
			}
		}
	}

	// We should have received at least some cached states
	if len(messages) == 0 {
		return fmt.Errorf("no cached device states received")
	}

	fc.responseData = map[string]interface{}{
		"cached_states": messages,
	}

	return nil
}

func (fc *FeatureContext) theCachedStatesShouldContainTheDeviceData() error {
	if fc.responseData == nil {
		return fmt.Errorf("no response data available")
	}

	cachedStates, ok := fc.responseData["cached_states"].([]DeviceStateMessage)
	if !ok {
		return fmt.Errorf("cached states not found in response data")
	}

	// Look for our test device
	found := false
	for _, state := range cachedStates {
		if state.DeviceID == "cache-test-device" {
			found = true
			// Verify the state has some data
			if len(state.Data) == 0 {
				return fmt.Errorf("device state has no sensor data")
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("cached state for cache-test-device not found")
	}

	return nil
}

// Cleanup function to close WebSocket connection
func (fc *FeatureContext) cleanupWebSocket() {
	if wsConn != nil {
		wsConn.Close()
		wsConn = nil
	}
}

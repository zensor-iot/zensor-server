package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"

	"github.com/gorilla/websocket"
)

func TestDeviceSpecificWebSocketController_HandleWebSocket(t *testing.T) {
	// Create mock dependencies
	broker := async.NewLocalBroker()
	stateCache := usecases.NewDeviceStateCacheService()

	// Create controller
	controller := NewDeviceSpecificWebSocketController(broker, stateCache)
	defer controller.Shutdown()

	// Create test server
	router := http.NewServeMux()
	controller.AddRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		deviceID       string
		expectedStatus int
	}{
		{
			name:           "valid device ID",
			deviceID:       "test-device-123",
			expectedStatus: http.StatusSwitchingProtocols,
		},
		{
			name:           "empty device ID",
			deviceID:       "   ", // Use whitespace instead of empty string
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := server.URL + "/ws/devices/" + tt.deviceID + "/messages"
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Add WebSocket headers
			req.Header.Set("Upgrade", "websocket")
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
			req.Header.Set("Sec-WebSocket-Version", "13")

			// Make request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestDeviceSpecificWebSocketController_MessageRouting(t *testing.T) {
	// Create mock dependencies
	broker := async.NewLocalBroker()
	stateCache := usecases.NewDeviceStateCacheService()

	// Create controller
	controller := NewDeviceSpecificWebSocketController(broker, stateCache)
	defer controller.Shutdown()

	// Create test server
	router := http.NewServeMux()
	controller.AddRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket for device1
	device1ID := "device-1"
	wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + device1ID + "/messages"
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	defer conn1.Close()

	// Connect to WebSocket for device2
	device2ID := "device-2"
	wsURL2 := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + device2ID + "/messages"
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	defer conn2.Close()

	// Wait a bit for connections to be established
	time.Sleep(100 * time.Millisecond)

	// Publish a message for device1
	envelop := dto.Envelop{
		EndDeviceIDs: dto.EndDeviceIDs{
			DeviceID: device1ID,
		},
		ReceivedAt: time.Now(),
		UplinkMessage: dto.UplinkMessage{
			DecodedPayload: map[string][]dto.SensorData{
				"temperature": {{Index: 0, Value: 25.5}},
			},
		},
	}

	brokerMsg := async.BrokerMessage{
		Event: "uplink",
		Value: envelop,
	}

	err = broker.Publish(context.Background(), "device_messages", brokerMsg)
	if err != nil {
		t.Fatalf("failed to publish message: %v", err)
	}

	// Wait for message to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if device1 received the message
	conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg DeviceSpecificMessage
	err = conn1.ReadJSON(&msg)
	if err != nil {
		t.Errorf("device1 should have received message: %v", err)
	}

	if msg.DeviceID != device1ID {
		t.Errorf("expected device ID %s, got %s", device1ID, msg.DeviceID)
	}

	// Check that device2 did NOT receive the message
	conn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var msg2 DeviceSpecificMessage
	err = conn2.ReadJSON(&msg2)
	if err == nil {
		t.Error("device2 should not have received message")
	}
}

func TestDeviceSpecificWebSocketController_CachedState(t *testing.T) {
	// Create mock dependencies
	broker := async.NewLocalBroker()
	stateCache := usecases.NewDeviceStateCacheService()

	// Set up cached state for a device
	deviceID := "test-device"
	sensorData := map[string][]dto.SensorData{
		"temperature": {{Index: 0, Value: 25.5}},
		"humidity":    {{Index: 0, Value: 60.0}},
	}
	stateCache.SetState(context.Background(), deviceID, sensorData)

	// Create controller
	controller := NewDeviceSpecificWebSocketController(broker, stateCache)
	defer controller.Shutdown()

	// Create test server
	router := http.NewServeMux()
	controller.AddRoutes(router)
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + deviceID + "/messages"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait for cached state to be sent
	time.Sleep(100 * time.Millisecond)

	// Check if cached state was received
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	var msg DeviceSpecificStateMessage
	err = conn.ReadJSON(&msg)
	if err != nil {
		t.Fatalf("failed to read cached state: %v", err)
	}

	if msg.DeviceID != deviceID {
		t.Errorf("expected device ID %s, got %s", deviceID, msg.DeviceID)
	}

	if msg.Type != "device_state" {
		t.Errorf("expected message type 'device_state', got %s", msg.Type)
	}

	// Check that sensor data is present
	if len(msg.Data["temperature"]) == 0 {
		t.Error("expected temperature data to be present")
	}

	if len(msg.Data["humidity"]) == 0 {
		t.Error("expected humidity data to be present")
	}
}

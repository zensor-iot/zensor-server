package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"

	"github.com/gorilla/websocket"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("DeviceSpecificWebSocketController", func() {
	var (
		broker     *async.LocalBroker
		stateCache usecases.DeviceStateCacheService
		controller *httpapi.DeviceSpecificWebSocketController
		router     *http.ServeMux
		server     *httptest.Server
	)

	ginkgo.BeforeEach(func() {
		// Create mock dependencies
		broker = async.NewLocalBroker()
		stateCache = persistence.NewSimpleDeviceStateCacheService()

		// Create controller
		controller = httpapi.NewDeviceSpecificWebSocketController(broker, stateCache)

		// Give the controller time to initialize
		time.Sleep(50 * time.Millisecond)

		// Create test server
		router = http.NewServeMux()
		controller.AddRoutes(router)
		server = httptest.NewServer(router)
	})

	ginkgo.AfterEach(func() {
		controller.Shutdown()
		// Give the controller time to shut down properly
		time.Sleep(100 * time.Millisecond)
		server.Close()
	})

	ginkgo.Context("HandleWebSocket", func() {
		ginkgo.When("handling WebSocket upgrade requests", func() {
			ginkgo.It("should accept valid device ID", func() {
				// Create request
				url := server.URL + "/ws/devices/test-device-123/messages"
				req, err := http.NewRequest("GET", url, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Add WebSocket headers
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
				req.Header.Set("Sec-WebSocket-Version", "13")

				// Make request
				resp, err := http.DefaultClient.Do(req)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer resp.Body.Close()

				// Check status code
				gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusSwitchingProtocols))
			})

			ginkgo.It("should reject empty device ID", func() {
				// Create request with whitespace device ID
				url := server.URL + "/ws/devices/   /messages"
				req, err := http.NewRequest("GET", url, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Add WebSocket headers
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
				req.Header.Set("Sec-WebSocket-Version", "13")

				// Make request
				resp, err := http.DefaultClient.Do(req)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer resp.Body.Close()

				// Check status code
				gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusBadRequest))
			})
		})
	})

	ginkgo.Context("MessageRouting", func() {
		ginkgo.When("publishing messages to specific devices", func() {
			ginkgo.It("should route messages only to the correct device", func() {
				// Connect to WebSocket for device1
				device1ID := "device-1"
				wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + device1ID + "/messages"
				conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer conn1.Close()

				// Connect to WebSocket for device2
				device2ID := "device-2"
				wsURL2 := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + device2ID + "/messages"
				conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
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
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Wait for message to be processed
				time.Sleep(100 * time.Millisecond)

				// Check if device1 received the message
				conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
				var msg httpapi.DeviceSpecificMessage
				err = conn1.ReadJSON(&msg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(msg.DeviceID).To(gomega.Equal(device1ID))

				// Check that device2 did NOT receive the message
				conn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				var msg2 httpapi.DeviceSpecificMessage
				err = conn2.ReadJSON(&msg2)
				gomega.Expect(err).To(gomega.HaveOccurred(), "device2 should not have received message")
			})
		})
	})

	ginkgo.Context("CachedState", func() {
		ginkgo.When("connecting with existing cached state", func() {
			ginkgo.It("should send cached state to the device", func() {
				// Set up cached state for a device
				deviceID := "test-device"
				sensorData := map[string][]dto.SensorData{
					"temperature": {{Index: 0, Value: 25.5}},
					"humidity":    {{Index: 0, Value: 60.0}},
				}
				stateCache.SetState(context.Background(), deviceID, sensorData)

				// Connect to WebSocket
				wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/devices/" + deviceID + "/messages"
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer conn.Close()

				// Wait for cached state to be sent
				time.Sleep(100 * time.Millisecond)

				// Check if cached state was received
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				var msg httpapi.DeviceSpecificStateMessage
				err = conn.ReadJSON(&msg)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(msg.DeviceID).To(gomega.Equal(deviceID))
				gomega.Expect(msg.Type).To(gomega.Equal("device_state"))

				// Check that sensor data is present
				gomega.Expect(msg.Data["temperature"]).NotTo(gomega.BeEmpty(), "expected temperature data to be present")
				gomega.Expect(msg.Data["humidity"]).NotTo(gomega.BeEmpty(), "expected humidity data to be present")
			})
		})
	})
})

package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"zensor-server/internal/infra/async"
	"zensor-server/internal/victron/dto"
	"zensor-server/internal/victron/httpapi"

	"github.com/gorilla/websocket"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("VictronWebSocketController", func() {
	var (
		broker     *async.LocalBroker
		controller *httpapi.VictronWebSocketController
		router     *http.ServeMux
		server     *httptest.Server
	)

	ginkgo.BeforeEach(func() {
		broker = async.NewLocalBroker()
		controller = httpapi.NewVictronWebSocketController(broker)
		time.Sleep(50 * time.Millisecond)

		router = http.NewServeMux()
		controller.AddRoutes(router)
		server = httptest.NewServer(router)
	})

	ginkgo.AfterEach(func() {
		controller.Shutdown()
		time.Sleep(100 * time.Millisecond)
		server.Close()
	})

	publishSystemTelemetry := func(broker *async.LocalBroker, path string, value float64) {
		telemetry := dto.VictronTelemetry{
			PortalID:    "d41243b4e8e4",
			ServiceType: "system",
			Instance:    0,
			Path:        path,
			Value:       dto.VictronValue{Value: value},
			Topic:       "N/d41243b4e8e4/system/0/" + path,
		}

		err := broker.Publish(context.Background(), async.BrokerTopicName("victron_data"), async.BrokerMessage{
			Event: "system_" + path,
			Value: telemetry,
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	ginkgo.Context("system service telemetry", func() {
		ginkgo.When("the GX device publishes its pre-aggregated system/0 topics", func() {
			ginkgo.It("should populate the system summary from those topics", func() {
				wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/victron/status"
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer conn.Close()

				time.Sleep(100 * time.Millisecond)

				// initial snapshot sent right after registration
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				var initial httpapi.VictronSystemStatusMessage
				gomega.Expect(conn.ReadJSON(&initial)).To(gomega.Succeed())

				publishSystemTelemetry(broker, "Ac/Grid/L1/Power", 500)
				publishSystemTelemetry(broker, "Ac/Consumption/L1/Power", 800)
				publishSystemTelemetry(broker, "Dc/Battery/Soc", 87)
				publishSystemTelemetry(broker, "Dc/Battery/Power", 300)
				publishSystemTelemetry(broker, "Dc/Pv/Power", 1200)

				var last httpapi.VictronSystemStatusMessage
				for range 5 {
					conn.SetReadDeadline(time.Now().Add(1 * time.Second))
					gomega.Expect(conn.ReadJSON(&last)).To(gomega.Succeed())
				}

				gomega.Expect(last.System.GridPower).To(gomega.Equal(500.0))
				gomega.Expect(last.System.AcLoadPower).To(gomega.Equal(800.0))
				gomega.Expect(last.System.BatterySOC).To(gomega.HaveValue(gomega.Equal(87.0)))
				gomega.Expect(last.System.BatteryPower).To(gomega.Equal(300.0))
				gomega.Expect(last.System.SolarPower).To(gomega.Equal(1200.0))
				gomega.Expect(last.System.IsCharging).To(gomega.BeTrue())
			})
		})

		ginkgo.When("no system topic has been received yet but a device reports its own power", func() {
			ginkgo.It("should not zero out the device-derived totals", func() {
				wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws/victron/status"
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer conn.Close()

				time.Sleep(100 * time.Millisecond)

				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				var initial httpapi.VictronSystemStatusMessage
				gomega.Expect(conn.ReadJSON(&initial)).To(gomega.Succeed())

				telemetry := dto.VictronTelemetry{
					PortalID:    "d41243b4e8e4",
					ServiceType: "acload",
					Instance:    0,
					Path:        "Ac/0/Power",
					Value:       dto.VictronValue{Value: 250},
					Topic:       "N/d41243b4e8e4/acload/0/Ac/0/Power",
				}
				err = broker.Publish(context.Background(), async.BrokerTopicName("victron_data"), async.BrokerMessage{
					Event: "acload_ac_0_power",
					Value: telemetry,
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				var last httpapi.VictronSystemStatusMessage
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				gomega.Expect(conn.ReadJSON(&last)).To(gomega.Succeed())

				gomega.Expect(last.System.AcLoadPower).To(gomega.Equal(250.0))
			})
		})
	})

	ginkgo.Context("concurrent websocket writes", func() {
		ginkgo.When("telemetry broadcasts while the keepalive ping fires", func() {
			ginkgo.It("should not panic from concurrent writes", func() {
				pingBroker := async.NewLocalBroker()
				pingController := httpapi.NewVictronWebSocketControllerWithPingInterval(pingBroker, 20*time.Millisecond)
				pingRouter := http.NewServeMux()
				pingController.AddRoutes(pingRouter)
				pingServer := httptest.NewServer(pingRouter)
				defer func() {
					pingController.Shutdown()
					time.Sleep(50 * time.Millisecond)
					pingServer.Close()
				}()
				time.Sleep(50 * time.Millisecond)

				wsURL := strings.Replace(pingServer.URL, "http", "ws", 1) + "/ws/victron/status"
				conns := make([]*websocket.Conn, 0, 2)
				for range 2 {
					conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					conns = append(conns, conn)
					defer conn.Close()
				}

				done := make(chan struct{})
				go func() {
					defer close(done)
					for range 30 {
						telemetry := dto.VictronTelemetry{
							PortalID:    "d41243b4e8e4",
							ServiceType: "system",
							Instance:    0,
							Path:        "Dc/Battery/Soc",
							Value:       dto.VictronValue{Value: 87},
							Topic:       "N/d41243b4e8e4/system/0/Dc/Battery/Soc",
						}
						_ = pingBroker.Publish(context.Background(), async.BrokerTopicName("victron_data"), async.BrokerMessage{
							Event: "system_battery_soc",
							Value: telemetry,
						})
						time.Sleep(5 * time.Millisecond)
					}
				}()

				for _, conn := range conns {
					go func(c *websocket.Conn) {
						for {
							if _, _, err := c.ReadMessage(); err != nil {
								return
							}
						}
					}(conn)
				}

				gomega.Eventually(done, 5*time.Second, 10*time.Millisecond).Should(gomega.BeClosed())
			})
		})
	})
})

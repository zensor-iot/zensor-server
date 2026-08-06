package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"zensor-server/internal/infra/async"
	"zensor-server/internal/victron/dto"
	"zensor-server/internal/victron/httpapi"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("VictronWebSocketController summary endpoint", func() {
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

	publishSystemTelemetry := func(path string, value float64) {
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

	ginkgo.Context("GET /v1/victron/summary", func() {
		ginkgo.When("telemetry has already been received", func() {
			ginkgo.BeforeEach(func() {
				publishSystemTelemetry("Ac/Grid/L1/Power", 500)
				publishSystemTelemetry("Ac/Consumption/L1/Power", 800)
				publishSystemTelemetry("Battery/Soc", 87)
				publishSystemTelemetry("Dc/Battery/Power", 300)
				publishSystemTelemetry("Dc/Pv/Power", 1200)
				time.Sleep(100 * time.Millisecond)
			})

			ginkgo.It("should return the grid, battery and consumption figures", func() {
				response, err := http.Get(server.URL + "/v1/victron/summary")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer response.Body.Close()

				gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))

				var body httpapi.VictronSystemSummary
				gomega.Expect(json.NewDecoder(response.Body).Decode(&body)).To(gomega.Succeed())

				gomega.Expect(body.GridPower).To(gomega.Equal(500.0))
				gomega.Expect(body.AcLoadPower).To(gomega.Equal(800.0))
				gomega.Expect(body.BatterySOC).To(gomega.Equal(87.0))
				gomega.Expect(body.BatteryPower).To(gomega.Equal(300.0))
				gomega.Expect(body.SolarPower).To(gomega.Equal(1200.0))
				gomega.Expect(body.IsCharging).To(gomega.BeTrue())
			})

			ginkgo.It("should not include the per-device snapshot", func() {
				response, err := http.Get(server.URL + "/v1/victron/summary")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer response.Body.Close()

				var body map[string]any
				gomega.Expect(json.NewDecoder(response.Body).Decode(&body)).To(gomega.Succeed())

				gomega.Expect(body).NotTo(gomega.HaveKey("data"))
				gomega.Expect(body).NotTo(gomega.HaveKey("raw"))
				gomega.Expect(body).To(gomega.HaveKey("grid_power"))
			})
		})

		ginkgo.When("no telemetry has been received yet", func() {
			ginkgo.It("should report the data as unavailable", func() {
				response, err := http.Get(server.URL + "/v1/victron/summary")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer response.Body.Close()

				gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusServiceUnavailable))
			})
		})
	})
})

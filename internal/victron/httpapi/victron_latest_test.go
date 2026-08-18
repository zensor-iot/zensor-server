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

var _ = ginkgo.Describe("VictronWebSocketController latest data endpoint", func() {
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

	publishTelemetry := func(serviceType, path string, value float64) {
		telemetry := dto.VictronTelemetry{
			PortalID:    "d41243b4e8e4",
			ServiceType: serviceType,
			Instance:    0,
			Path:        path,
			Value:       dto.VictronValue{Value: value},
			Topic:       "N/d41243b4e8e4/" + serviceType + "/0/" + path,
		}

		err := broker.Publish(context.Background(), async.BrokerTopicName("victron_data"), async.BrokerMessage{
			Event: serviceType + "_" + path,
			Value: telemetry,
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}

	ginkgo.Context("GET /v1/victron/latest", func() {
		ginkgo.When("telemetry has already been received", func() {
			ginkgo.BeforeEach(func() {
				publishTelemetry("system", "Dc/Battery/Soc", 87)
				publishTelemetry("system", "Dc/Pv/Power", 1200)
				time.Sleep(100 * time.Millisecond)
			})

			ginkgo.It("should return the latest snapshot with its system summary", func() {
				response, err := http.Get(server.URL + "/v1/victron/latest")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer response.Body.Close()

				gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))

				var body httpapi.VictronSystemStatusMessage
				gomega.Expect(json.NewDecoder(response.Body).Decode(&body)).To(gomega.Succeed())

				gomega.Expect(body.Type).To(gomega.Equal("victron_status"))
				gomega.Expect(body.Data.PortalID).To(gomega.Equal("d41243b4e8e4"))
				gomega.Expect(body.System.BatterySOC).To(gomega.HaveValue(gomega.Equal(87.0)))
				gomega.Expect(body.System.SolarPower).To(gomega.Equal(1200.0))
			})
		})

		ginkgo.When("no telemetry has been received yet", func() {
			ginkgo.It("should report the data as unavailable", func() {
				response, err := http.Get(server.URL + "/v1/victron/latest")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer response.Body.Close()

				gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusServiceUnavailable))
			})
		})
	})
})

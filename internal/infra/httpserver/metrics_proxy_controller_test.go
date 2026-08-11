package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MetricsProxyController", func() {
	var (
		vmServer *httptest.Server
		proxied  *MetricsProxyController
		router   *http.ServeMux
		received string
	)

	ginkgo.BeforeEach(func() {
		received = ""
		vmServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"status":"success","data":{"resultType":"matrix","result":[]}}`)
		}))

		proxied = NewMetricsProxyController(vmServer.URL)
		router = http.NewServeMux()
		proxied.AddRoutes(router)
	})

	ginkgo.AfterEach(func() {
		vmServer.Close()
	})

	ginkgo.When("a query_range request hits /v1/metrics/", func() {
		ginkgo.It("should forward it to the VictoriaMetrics api and return the payload", func() {
			req := httptest.NewRequest(http.MethodGet,
				"/v1/metrics/query_range?query=zensor_server_victron_battery_soc&start=1700000000&end=1700003600&step=60s",
				nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			gomega.Expect(received).To(gomega.Equal("/api/v1/query_range"))
			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

			var body map[string]any
			gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
			gomega.Expect(body["status"]).To(gomega.Equal("success"))
		})
	})

	ginkgo.When("a series request hits /v1/metrics/", func() {
		ginkgo.It("should map it onto /api/v1/", func() {
			req := httptest.NewRequest(http.MethodGet, "/v1/metrics/series?match[]=zensor_server_victron_battery_soc", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			gomega.Expect(received).To(gomega.Equal("/api/v1/series"))
			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			gomega.Expect(strings.Contains(rec.Body.String(), "result")).To(gomega.BeTrue())
		})
	})

	ginkgo.When("the proxied request fails", func() {
		ginkgo.It("should propagate a 502 response", func() {
			vmServer.Close()

			req := httptest.NewRequest(http.MethodGet,
				"/v1/metrics/query_range?query=zensor_server_victron_battery_soc", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadGateway))
		})
	})
})

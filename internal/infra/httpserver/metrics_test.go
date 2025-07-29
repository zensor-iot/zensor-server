package httpserver

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
)

var _ = ginkgo.Describe("Metrics", func() {
	ginkgo.Context("MetricsMiddleware", func() {
		ginkgo.When("using metrics middleware", func() {
			ginkgo.It("should collect metrics correctly", func() {
				// Set up a test meter provider
				reader := metric.NewManualReader()
				provider := metric.NewMeterProvider(metric.WithReader(reader))
				otel.SetMeterProvider(provider)

				// Reset metrics initialization for testing
				ResetMetricsForTesting()

				// Create a test handler that simulates some work
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate some processing time
					time.Sleep(10 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test response"))
				})

				// Create middleware
				middleware := MetricsMiddleware()
				handler := middleware(testHandler)

				// Create test request
				req := httptest.NewRequest("GET", "/test/endpoint", nil)
				w := httptest.NewRecorder()

				// Execute request
				handler.ServeHTTP(w, req)

				// Verify response
				gomega.Expect(w.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(w.Body.String()).To(gomega.Equal("test response"))

				// Verify that metrics were initialized
				gomega.Expect(IsMetricsInitialized()).To(gomega.BeTrue())
			})
		})
	})

	ginkgo.Context("ExtractEndpoint", func() {
		var (
			path     string
			expected string
		)

		ginkgo.When("extracting endpoint from path", func() {
			ginkgo.It("should handle root path", func() {
				path = "/"
				expected = "root"

				result := extractEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle simple endpoint", func() {
				path = "/api"
				expected = "api"

				result := extractEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle nested endpoint", func() {
				path = "/api/v1/users"
				expected = "api"

				result := extractEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle empty path", func() {
				path = ""
				expected = "root"

				result := extractEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle single segment", func() {
				path = "/healthz"
				expected = "healthz"

				result := extractEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})
		})
	})

	ginkgo.Context("ResponseWriter", func() {
		var (
			recorder      *httptest.ResponseRecorder
			wrappedWriter *responseWriter
		)

		ginkgo.When("using response writer wrapper", func() {
			ginkgo.BeforeEach(func() {
				// Create a test response writer
				recorder = httptest.NewRecorder()
				wrappedWriter = &responseWriter{ResponseWriter: recorder, statusCode: http.StatusOK}
			})

			ginkgo.It("should handle WriteHeader correctly", func() {
				wrappedWriter.WriteHeader(http.StatusNotFound)
				gomega.Expect(wrappedWriter.statusCode).To(gomega.Equal(http.StatusNotFound))
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
			})

			ginkgo.It("should handle Write correctly", func() {
				_, err := wrappedWriter.Write([]byte("test"))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(recorder.Body.String()).To(gomega.Equal("test"))
			})
		})
	})

	ginkgo.Context("ResponseWriterHijacker", func() {
		var (
			recorder      *httptest.ResponseRecorder
			wrappedWriter *responseWriter
		)

		ginkgo.When("testing hijacker interface", func() {
			ginkgo.BeforeEach(func() {
				// Create a test response writer that implements http.Hijacker
				recorder = httptest.NewRecorder()
				wrappedWriter = &responseWriter{ResponseWriter: recorder, statusCode: http.StatusOK}
			})

			ginkgo.It("should implement http.Hijacker interface", func() {
				// Test that our wrapper implements http.Hijacker interface
				_, isHijacker := interface{}(wrappedWriter).(http.Hijacker)
				gomega.Expect(isHijacker).To(gomega.BeTrue())
			})

			ginkgo.It("should return error when hijacking is not supported", func() {
				// Test calling Hijack (it should return an error since httptest.ResponseRecorder doesn't support hijacking)
				_, _, err := wrappedWriter.Hijack()
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("underlying ResponseWriter does not support hijacking"))
			})
		})
	})
})

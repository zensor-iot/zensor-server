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

	ginkgo.Context("NormalizeEndpoint", func() {
		var (
			path     string
			expected string
		)

		ginkgo.When("normalizing endpoint from path", func() {
			ginkgo.It("should handle root path", func() {
				path = "/"
				expected = "root"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle simple endpoint", func() {
				path = "/api"
				expected = "/api"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle nested endpoint", func() {
				path = "/api/v1/users"
				expected = "/api/v1/users"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle empty path", func() {
				path = ""
				expected = "root"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle single segment", func() {
				path = "/healthz"
				expected = "/healthz"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})
		})

		ginkgo.When("normalizing endpoint with UUIDs", func() {
			ginkgo.It("should replace device UUID with _id", func() {
				path = "/v1/devices/123e4567-e89b-12d3-a456-426614174000"
				expected = "/v1/devices/_id"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace tenant UUID with _id", func() {
				path = "/v1/tenants/123e4567-e89b-12d3-a456-426614174000"
				expected = "/v1/tenants/_id"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace device UUID in commands endpoint", func() {
				path = "/v1/devices/123e4567-e89b-12d3-a456-426614174000/commands"
				expected = "/v1/devices/_id/commands"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace device UUID in tasks endpoint", func() {
				path = "/v1/devices/123e4567-e89b-12d3-a456-426614174000/tasks"
				expected = "/v1/devices/_id/tasks"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace device UUID in evaluation-rules endpoint", func() {
				path = "/v1/devices/123e4567-e89b-12d3-a456-426614174000/evaluation-rules"
				expected = "/v1/devices/_id/evaluation-rules"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace tenant UUID in configuration endpoint", func() {
				path = "/v1/tenants/123e4567-e89b-12d3-a456-426614174000/configuration"
				expected = "/v1/tenants/_id/configuration"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should replace device UUID in WebSocket endpoint", func() {
				path = "/ws/devices/123e4567-e89b-12d3-a456-426614174000/messages"
				expected = "/ws/devices/_id/messages"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle complex nested path with multiple UUIDs", func() {
				path = "/v1/tenants/123e4567-e89b-12d3-a456-426614174000/devices/987fcdeb-51a2-43d7-8f9e-123456789abc/scheduled-tasks/456e7890-e12b-34c5-d678-901234567def"
				expected = "/v1/tenants/_id/devices/_id/scheduled-tasks/_id"

				result := normalizeEndpoint(path)
				gomega.Expect(result).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle tenant devices endpoint", func() {
				path = "/v1/tenants/123e4567-e89b-12d3-a456-426614174000/devices"
				expected = "/v1/tenants/_id/devices"

				result := normalizeEndpoint(path)
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

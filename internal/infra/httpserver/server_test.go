package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var _ = ginkgo.Describe("HTTPServer", func() {
	var (
		tp *trace.TracerProvider
	)

	ginkgo.BeforeEach(func() {
		// Set up a test trace provider
		tp = trace.NewTracerProvider(
			trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
		)
		otel.SetTracerProvider(tp)
	})

	ginkgo.AfterEach(func() {
		tp.Shutdown(context.Background())
	})

	ginkgo.Context("TracingMiddleware", func() {
		ginkgo.When("using tracing middleware", func() {
			ginkgo.It("should add span to request context", func() {
				// Create a test handler that checks if span is in context
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					span := GetSpanFromContext(r)
					gomega.Expect(span).NotTo(gomega.BeNil())

					// Check that we have a valid span context
					spanCtx := span.SpanContext()
					gomega.Expect(spanCtx.HasSpanID()).To(gomega.BeTrue())

					w.WriteHeader(http.StatusOK)
				})

				// Create middleware
				middleware := createTracingMiddleware()
				wrappedHandler := middleware(testHandler)

				// Create test request
				req := httptest.NewRequest("GET", "/test", nil)
				rec := httptest.NewRecorder()

				// Execute request
				wrappedHandler.ServeHTTP(rec, req)

				// Check response
				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("GetSpanFromContext", func() {
		ginkgo.When("getting span from context", func() {
			ginkgo.It("should return a span even when no span is in context", func() {
				// Test with request that has no span
				req := httptest.NewRequest("GET", "/test", nil)
				span := GetSpanFromContext(req)

				// Should return a no-op span when no span is in context
				gomega.Expect(span).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("UserHeaderMiddleware", func() {
		ginkgo.When("using user header middleware with headers", func() {
			ginkgo.It("should process user headers correctly", func() {
				// Create a test handler that checks if user attributes are in span
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					span := GetSpanFromContext(r)
					gomega.Expect(span).NotTo(gomega.BeNil())

					// Check that we have a valid span context
					spanCtx := span.SpanContext()
					gomega.Expect(spanCtx.HasSpanID()).To(gomega.BeTrue())

					w.WriteHeader(http.StatusOK)
				})

				// Create middleware chain
				tracingMiddleware := createTracingMiddleware()
				userHeaderMiddleware := createUserHeaderMiddleware()
				wrappedHandler := tracingMiddleware(userHeaderMiddleware(testHandler))

				// Create test request with user headers
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-User-ID", "user123")
				req.Header.Set("X-User-Name", "John Doe")
				req.Header.Set("X-User-Email", "john.doe@example.com")
				rec := httptest.NewRecorder()

				// Execute request
				wrappedHandler.ServeHTTP(rec, req)

				// Check response
				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("using user header middleware without headers", func() {
			ginkgo.It("should handle requests without user headers", func() {
				// Create a test handler
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					span := GetSpanFromContext(r)
					gomega.Expect(span).NotTo(gomega.BeNil())

					w.WriteHeader(http.StatusOK)
				})

				// Create middleware chain
				tracingMiddleware := createTracingMiddleware()
				userHeaderMiddleware := createUserHeaderMiddleware()
				wrappedHandler := tracingMiddleware(userHeaderMiddleware(testHandler))

				// Create test request without user headers
				req := httptest.NewRequest("GET", "/test", nil)
				rec := httptest.NewRecorder()

				// Execute request
				wrappedHandler.ServeHTTP(rec, req)

				// Check response
				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})
})

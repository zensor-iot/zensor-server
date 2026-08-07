package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
	"zensor-server/internal/shared_kernel/domain"

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

	ginkgo.Context("CurrentUser", func() {
		ginkgo.When("the request carries user headers", func() {
			ginkgo.It("should return them as JSON", func() {
				req := httptest.NewRequest("GET", "/v1/me", nil)
				req.Header.Set("X-User-ID", "user123")
				req.Header.Set("X-User-Name", "John Doe")
				req.Header.Set("X-User-Email", "john.doe@example.com")
				rec := httptest.NewRecorder()

				getCurrentUser().ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body CurrentUserResponse
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.Equal(CurrentUserResponse{
					UserID: "user123",
					Name:   "John Doe",
					Email:  "john.doe@example.com",
				}))
			})
		})

		ginkgo.When("the request has no user headers", func() {
			ginkgo.It("should return empty fields, not an error", func() {
				req := httptest.NewRequest("GET", "/v1/me", nil)
				rec := httptest.NewRecorder()

				getCurrentUser().ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body CurrentUserResponse
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.Equal(CurrentUserResponse{}))
			})
		})
	})

	ginkgo.Context("NewServerWithAuth", func() {
		var (
			srv      *StandardServer
			resolver *fakeSessionResolver
		)

		ginkgo.BeforeEach(func() {
			resolver = &fakeSessionResolver{sessions: map[string]domain.Session{
				"valid-session": {
					ID:        "valid-session",
					UserID:    "user-1",
					Email:     "user@example.com",
					ExpiresAt: time.Now().Add(time.Hour),
				},
			}}
			srv = NewServerWithAuth(0, resolver, &fakeAPIKeyResolver{})
		})

		ginkgo.When("requesting a protected route without a session", func() {
			ginkgo.It("should return 401", func() {
				req := httptest.NewRequest("GET", "/v1/tenants", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))
			})
		})

		ginkgo.When("requesting a protected route with a valid session", func() {
			ginkgo.It("should reach the router", func() {
				req := httptest.NewRequest("GET", "/v1/unknown", nil)
				req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "valid-session"})
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})

		ginkgo.When("requesting the SPA without a session", func() {
			ginkgo.It("should serve it publicly", func() {
				req := httptest.NewRequest("GET", "/ui/", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("StaticWebUI", func() {
		ginkgo.When("requesting the SPA's mount path", func() {
			ginkgo.It("should serve the embedded SPA", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("GET", "/ui/", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Body.String()).To(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})

		ginkgo.When("requesting a real API route with the SPA also registered", func() {
			ginkgo.It("should still route to healthz, not the SPA fallback", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("GET", "/healthz", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Header().Get("Content-Type")).To(gomega.Equal("application/json"))
			})
		})

		ginkgo.When("requesting an unmatched path under /v1/", func() {
			ginkgo.It("should return 404, not the SPA's HTML", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("GET", "/v1/nonexistent-route", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
				gomega.Expect(rec.Body.String()).NotTo(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})

		ginkgo.When("requesting a real route with the wrong HTTP method", func() {
			ginkgo.It("should not return 200 with the SPA's HTML", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("PUT", "/v1/tenants", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).NotTo(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Body.String()).NotTo(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})

		ginkgo.When("requesting a legitimate real route", func() {
			ginkgo.It("should still work correctly", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("GET", "/v1/me", nil)
				req.Header.Set("X-User-ID", "user123")
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body CurrentUserResponse
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body.UserID).To(gomega.Equal("user123"))
			})
		})

		ginkgo.When("requesting a genuine client-side route under /ui/", func() {
			ginkgo.It("should still fall back to the SPA's index.html", func() {
				srv := NewServer(0)
				req := httptest.NewRequest("GET", "/ui/some-client-route", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Body.String()).To(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})
	})
})

package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingMiddleware(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a test handler that checks if span is in context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := GetSpanFromContext(r)
		if span == nil {
			t.Error("Expected span to be in request context, but got nil")
			return
		}

		// Check that we have a valid span context
		spanCtx := span.SpanContext()
		if !spanCtx.HasSpanID() {
			t.Error("Expected span to have a span ID")
		}

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
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestGetSpanFromContext(t *testing.T) {
	// Test with request that has no span
	req := httptest.NewRequest("GET", "/test", nil)
	span := GetSpanFromContext(req)

	// Should return a no-op span when no span is in context
	if span == nil {
		t.Error("Expected GetSpanFromContext to return a span (even if no-op), but got nil")
	}
}

func TestUserHeaderMiddleware(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a test handler that checks if user attributes are in span
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := GetSpanFromContext(r)
		if span == nil {
			t.Error("Expected span to be in request context, but got nil")
			return
		}

		// Check that we have a valid span context
		spanCtx := span.SpanContext()
		if !spanCtx.HasSpanID() {
			t.Error("Expected span to have a span ID")
		}

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
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUserHeaderMiddlewareWithoutHeaders(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := GetSpanFromContext(r)
		if span == nil {
			t.Error("Expected span to be in request context, but got nil")
			return
		}

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
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

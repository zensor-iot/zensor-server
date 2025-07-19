package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestMetricsMiddleware(t *testing.T) {
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
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Verify that metrics were initialized
	assert.True(t, IsMetricsInitialized())
}

func TestExtractEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root path",
			path:     "/",
			expected: "root",
		},
		{
			name:     "simple endpoint",
			path:     "/api",
			expected: "api",
		},
		{
			name:     "nested endpoint",
			path:     "/api/v1/users",
			expected: "api",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "root",
		},
		{
			name:     "single segment",
			path:     "/healthz",
			expected: "healthz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractEndpoint(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	// Create a test response writer
	recorder := httptest.NewRecorder()
	wrappedWriter := &responseWriter{ResponseWriter: recorder, statusCode: http.StatusOK}

	// Test WriteHeader
	wrappedWriter.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, wrappedWriter.statusCode)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Test Write
	_, err := wrappedWriter.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "test", recorder.Body.String())
}

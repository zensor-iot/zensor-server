package httpserver

import (
	"net/http"
	"time"
)

// ExampleMetricsUsage demonstrates how the metrics middleware automatically collects HTTP metrics
func ExampleMetricsUsage() {
	// This example shows how the metrics middleware automatically collects metrics
	// for all HTTP requests without any additional code needed in your handlers

	// Example handler that will automatically have metrics collected
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(100 * time.Millisecond)

		// Your business logic here
		// The metrics middleware will automatically:
		// 1. Record request duration
		// 2. Increment total request counter
		// 3. Track active requests
		// 4. Capture HTTP method, endpoint, and status code

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}

	// Example handler that returns an error status
	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		// Simulate an error condition
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}

	// Example handler for different HTTP methods
	createHandler := func(w http.ResponseWriter, r *http.Request) {
		// Simulate creating a resource
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "123", "status": "created"}`))
	}

	_ = handler       // Suppress unused variable warning
	_ = errorHandler  // Suppress unused variable warning
	_ = createHandler // Suppress unused variable warning
}

// ExampleMetricsData shows what metrics data looks like
func ExampleMetricsData() {
	/*
		The metrics middleware automatically collects the following metrics:

		1. http_request_duration_seconds (Histogram)
		   - Measures request processing time
		   - Buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 seconds
		   - Attributes: http.method, http.endpoint, http.status_code

		2. http_requests_total (Counter)
		   - Counts total number of requests
		   - Attributes: http.method, http.endpoint, http.status_code

		3. http_requests_active (UpDownCounter)
		   - Tracks currently active requests
		   - Attributes: http.method, http.endpoint

		Example metric data:
		http_request_duration_seconds{http_method="GET",http_endpoint="api",http_status_code="200"} 0.123
		http_requests_total{http_method="GET",http_endpoint="api",http_status_code="200"} 150
		http_requests_active{http_method="GET",http_endpoint="api"} 5
	*/
}

// ExampleEndpointExtraction shows how endpoints are extracted from URLs
func ExampleEndpointExtraction() {
	/*
		The metrics middleware extracts endpoint names from URL paths:

		URL Path                    -> Endpoint Name
		/                          -> "root"
		/api                       -> "api"
		/api/v1/users              -> "api"
		/healthz                   -> "healthz"
		/metrics                   -> "metrics"
		/devices/123/commands      -> "devices"

		This helps group similar endpoints together for better metrics aggregation.
	*/
}

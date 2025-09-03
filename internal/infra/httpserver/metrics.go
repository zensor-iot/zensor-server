package httpserver

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// HTTP metrics
	httpRequestDuration metric.Float64Histogram
	httpRequestTotal    metric.Int64Counter
	httpRequestActive   metric.Int64UpDownCounter
	metricsInitialized  bool
	metricsMutex        sync.Mutex

	// UUID regex pattern for identifying UUIDs in paths
	uuidRegex = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

// ResetMetricsForTesting resets the metrics initialization state for testing purposes
func ResetMetricsForTesting() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	metricsInitialized = false
}

// IsMetricsInitialized returns whether metrics have been initialized (for testing)
func IsMetricsInitialized() bool {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()
	return metricsInitialized
}

func initMetrics() {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	if metricsInitialized {
		return
	}

	meter := otel.GetMeterProvider().Meter("zensor-server")

	// Initialize HTTP request duration histogram
	var err error
	httpRequestDuration, err = meter.Float64Histogram(
		fmt.Sprintf("%s.%s", "zensor_server", "http.request.duration.seconds"),
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		panic(err)
	}

	// Initialize HTTP request total counter
	httpRequestTotal, err = meter.Int64Counter(
		fmt.Sprintf("%s.%s", "zensor_server", "http.requests.total"),
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		panic(err)
	}

	// Initialize HTTP request active up-down counter
	httpRequestActive, err = meter.Int64UpDownCounter(
		fmt.Sprintf("%s.%s", "zensor_server", "http.requests.active"),
		metric.WithDescription("Number of HTTP requests currently being processed"),
	)
	if err != nil {
		panic(err)
	}

	metricsInitialized = true
}

// MetricsMiddleware creates a middleware that measures HTTP request metrics
func MetricsMiddleware() func(http.Handler) http.Handler {
	// Initialize metrics if not already done
	initMetrics()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Normalize endpoint by replacing UUIDs with route patterns
			endpoint := normalizeEndpoint(r.URL.Path)

			// Increment active requests counter
			httpRequestActive.Add(r.Context(), 1,
				metric.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.endpoint", endpoint),
				),
			)

			// Create a response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(wrappedWriter, r)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Record metrics
			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.endpoint", endpoint),
				attribute.Int("http.status_code", wrappedWriter.statusCode),
			}

			// Record request duration
			httpRequestDuration.Record(r.Context(), duration, metric.WithAttributes(attrs...))

			// Increment total requests counter
			httpRequestTotal.Add(r.Context(), 1, metric.WithAttributes(attrs...))

			// Decrement active requests counter
			httpRequestActive.Add(r.Context(), -1,
				metric.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.endpoint", endpoint),
				),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Hijack implements http.Hijacker interface for websocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
}

// normalizeEndpoint normalizes the request path by replacing UUIDs with route patterns
// This helps group similar endpoints together for better metrics aggregation
func normalizeEndpoint(path string) string {
	if path == "" || path == "/" {
		return "root"
	}

	normalizedPath := uuidRegex.ReplaceAllStringFunc(path, func(uuid string) string {
		return "_id"
	})

	return normalizedPath
}

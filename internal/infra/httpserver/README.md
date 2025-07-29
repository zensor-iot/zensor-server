# HTTP Server with OpenTelemetry Tracing and Metrics

This package provides an HTTP server with built-in OpenTelemetry tracing and metrics middleware.

## Features

- **Automatic Request Tracing**: Every HTTP request automatically creates a new span with request details
- **Automatic Request Metrics**: HTTP request metrics are automatically collected and exposed
- **Context Propagation**: Spans are automatically added to the request context
- **Helper Functions**: Easy access to spans from request context
- **Custom Attributes**: Controllers can add custom attributes and events to spans

## Usage

### Basic Setup

The tracing and metrics middleware are automatically applied to all requests when using `httpserver.NewServer()`:

```go
server := httpserver.NewServer(controllers...)
```

### Accessing Spans in Controllers

```go
func (c *MyController) handleRequest() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the current span from the request context
        span := httpserver.GetSpanFromContext(r)
        
        // Add custom attributes
        span.SetAttributes(attribute.String("user.id", userID))
        span.SetAttributes(attribute.Int("items.count", len(items)))
        
        // Record errors
        if err != nil {
            span.RecordError(err)
        }
        
        // Your handler logic here...
    }
}
```

### User Header Validation

The server automatically validates and captures user information from request headers:

- `X-User-ID`: User identifier
- `X-User-Name`: User display name  
- `X-User-Email`: User email address

When present, these headers are automatically added as span attributes for tracing and observability.

**Middleware Order**: The user header middleware runs after the tracing middleware to ensure it can access the span created by the tracing middleware.

### Span Attributes

The tracing middleware automatically adds the following attributes to each request span:

- `http.method`: The HTTP method (GET, POST, etc.)
- `http.url`: The full request URL
- `http.user_agent`: The User-Agent header
- `http.remote_addr`: The client's IP address
- `user.id`: User identifier (when X-User-ID header is present)
- `user.name`: User display name (when X-User-Name header is present)
- `user.email`: User email address (when X-User-Email header is present)

### Metrics

The metrics middleware automatically collects the following HTTP metrics:

- **http_request_duration_seconds**: Histogram of request duration in seconds
- **http_requests_total**: Counter of total requests with attributes:
  - `http.method`: HTTP method (GET, POST, etc.)
  - `http.endpoint`: Endpoint name (extracted from URL path)
  - `http.status_code`: HTTP response status code
- **http_requests_active**: Up-down counter of currently active requests with attributes:
  - `http.method`: HTTP method
  - `http.endpoint`: Endpoint name

#### Metric Attributes

- **http.method**: The HTTP method (GET, POST, PUT, DELETE, etc.)
- **http.endpoint**: The endpoint name extracted from the URL path (e.g., "api", "healthz", "metrics")
- **http.status_code**: The HTTP response status code (200, 404, 500, etc.)

### Custom Attributes

You can add custom attributes to spans using the OpenTelemetry attribute package:

```go
import "go.opentelemetry.io/otel/attribute"

span.SetAttributes(
    attribute.String("device.id", deviceID),
    attribute.Int("commands.count", len(commands)),
    attribute.Bool("feature.enabled", true),
)
```

### Error Recording

Record errors in spans to maintain trace context:

```go
if err != nil {
    span.RecordError(err)
    span.SetAttributes(attribute.String("error.type", "validation_error"))
    // Handle error...
}
```

## Configuration

The tracing and metrics middleware use the global OpenTelemetry tracer and meter providers configured in your application. Make sure to initialize OpenTelemetry in your main function:

```go
// In main.go
shutdownOtel := startOTel()
defer shutdownOtel()
```

## Testing

The package includes tests that demonstrate proper usage of the tracing and metrics middleware. Tests use test providers to verify span creation, context propagation, and metrics collection.

## Dependencies

- `go.opentelemetry.io/otel`: Core OpenTelemetry functionality
- `go.opentelemetry.io/otel/trace`: Tracing interfaces
- `go.opentelemetry.io/otel/metric`: Metrics interfaces
- `go.opentelemetry.io/otel/attribute`: Span attributes
- `go.opentelemetry.io/otel/sdk/trace`: Trace provider implementation
- `go.opentelemetry.io/otel/sdk/metric`: Metrics provider implementation 
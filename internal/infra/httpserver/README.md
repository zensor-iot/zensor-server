# HTTP Server with OpenTelemetry Tracing

This package provides an HTTP server with built-in OpenTelemetry tracing middleware.

## Features

- **Automatic Request Tracing**: Every HTTP request automatically creates a new span with request details
- **Context Propagation**: Spans are automatically added to the request context
- **Helper Functions**: Easy access to spans from request context
- **Custom Attributes**: Controllers can add custom attributes and events to spans

## Usage

### Basic Setup

The tracing middleware is automatically applied to all requests when using `httpserver.NewServer()`:

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

### Span Attributes

The middleware automatically adds the following attributes to each request span:

- `http.method`: The HTTP method (GET, POST, etc.)
- `http.url`: The full request URL
- `http.user_agent`: The User-Agent header
- `http.remote_addr`: The client's IP address

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

The tracing middleware uses the global OpenTelemetry tracer provider configured in your application. Make sure to initialize OpenTelemetry tracing in your main function:

```go
// In main.go
shutdownOtel := startOTel()
defer shutdownOtel()
```

## Testing

The package includes tests that demonstrate proper usage of the tracing middleware. Tests use a test trace provider to verify span creation and context propagation.

## Dependencies

- `go.opentelemetry.io/otel`: Core OpenTelemetry functionality
- `go.opentelemetry.io/otel/trace`: Tracing interfaces
- `go.opentelemetry.io/otel/attribute`: Span attributes
- `go.opentelemetry.io/otel/sdk/trace`: Trace provider implementation 
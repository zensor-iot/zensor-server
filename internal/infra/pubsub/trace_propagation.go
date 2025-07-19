package pubsub

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/trace"
)

// TraceHeaders contains the OpenTelemetry trace context headers
type TraceHeaders struct {
	TraceID    string `json:"trace_id"`
	SpanID     string `json:"span_id"`
	TraceFlags string `json:"trace_flags"`
}

// ExtractTraceFromContext extracts trace context from the given context
// and returns it as TraceHeaders for serialization
func ExtractTraceFromContext(ctx context.Context) TraceHeaders {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	return TraceHeaders{
		TraceID:    spanCtx.TraceID().String(),
		SpanID:     spanCtx.SpanID().String(),
		TraceFlags: strconv.FormatUint(uint64(spanCtx.TraceFlags()), 16),
	}
}

// InjectTraceIntoContext creates a new context with the trace context
// from the provided TraceHeaders
func InjectTraceIntoContext(ctx context.Context, headers TraceHeaders) context.Context {
	if headers.TraceID == "" || headers.SpanID == "" {
		// No trace context available, return original context
		return ctx
	}

	traceID, err := trace.TraceIDFromHex(headers.TraceID)
	if err != nil {
		// Invalid trace ID, return original context
		return ctx
	}

	spanID, err := trace.SpanIDFromHex(headers.SpanID)
	if err != nil {
		// Invalid span ID, return original context
		return ctx
	}

	var traceFlags trace.TraceFlags
	if headers.TraceFlags != "" {
		if flags, err := strconv.ParseUint(headers.TraceFlags, 16, 8); err == nil {
			traceFlags = trace.TraceFlags(flags)
		}
	}

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
		Remote:     true, // Mark as remote since it came from Kafka
	})

	return trace.ContextWithSpanContext(ctx, spanCtx)
}

// SerializeTraceHeaders converts TraceHeaders to Kafka headers
func SerializeTraceHeaders(headers TraceHeaders) map[string][]byte {
	return map[string][]byte{
		"trace_id":    []byte(headers.TraceID),
		"span_id":     []byte(headers.SpanID),
		"trace_flags": []byte(headers.TraceFlags),
	}
}

// DeserializeTraceHeaders converts Kafka headers back to TraceHeaders
func DeserializeTraceHeaders(headers map[string][]byte) TraceHeaders {
	traceHeaders := TraceHeaders{}

	if traceID, exists := headers["trace_id"]; exists {
		traceHeaders.TraceID = string(traceID)
	}
	if spanID, exists := headers["span_id"]; exists {
		traceHeaders.SpanID = string(spanID)
	}
	if traceFlags, exists := headers["trace_flags"]; exists {
		traceHeaders.TraceFlags = string(traceFlags)
	}

	return traceHeaders
}

// ExtractTraceFromKafkaHeaders extracts trace context from Kafka headers
// and injects it into the provided context
func ExtractTraceFromKafkaHeaders(ctx context.Context, headers map[string][]byte) context.Context {
	traceHeaders := DeserializeTraceHeaders(headers)
	return InjectTraceIntoContext(ctx, traceHeaders)
}

// CreateChildSpan creates a new span as a child of the span in the context
// This is useful for creating new spans in message handlers
func CreateChildSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("kafka-consumer")
	return tracer.Start(ctx, name, opts...)
}

package pubsub

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTracePropagation(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a test context with a span
	ctx, span := tp.Tracer("test").Start(context.Background(), "test.span")
	defer span.End()

	// Test extracting trace headers
	traceHeaders := ExtractTraceFromContext(ctx)

	if traceHeaders.TraceID == "" {
		t.Error("Expected trace ID to be set")
	}

	if traceHeaders.SpanID == "" {
		t.Error("Expected span ID to be set")
	}

	// Test serializing to Kafka headers
	kafkaHeaders := SerializeTraceHeaders(traceHeaders)

	if len(kafkaHeaders) == 0 {
		t.Error("Expected Kafka headers to be created")
	}

	if string(kafkaHeaders["trace_id"]) != traceHeaders.TraceID {
		t.Error("Expected trace ID to match in headers")
	}

	if string(kafkaHeaders["span_id"]) != traceHeaders.SpanID {
		t.Error("Expected span ID to match in headers")
	}

	// Test deserializing from Kafka headers
	deserializedHeaders := DeserializeTraceHeaders(kafkaHeaders)

	if deserializedHeaders.TraceID != traceHeaders.TraceID {
		t.Error("Expected deserialized trace ID to match")
	}

	if deserializedHeaders.SpanID != traceHeaders.SpanID {
		t.Error("Expected deserialized span ID to match")
	}

	// Test injecting trace into context
	newCtx := InjectTraceIntoContext(context.Background(), deserializedHeaders)

	// Check that trace context was injected
	newSpan := oteltrace.SpanFromContext(newCtx)
	if !newSpan.SpanContext().HasSpanID() {
		t.Error("Expected new context to have span ID")
	}

	// Verify trace IDs match
	if newSpan.SpanContext().TraceID() != span.SpanContext().TraceID() {
		t.Error("Expected trace IDs to match")
	}
}

func TestTracePropagationWithEmptyHeaders(t *testing.T) {
	// Test with empty headers
	emptyHeaders := TraceHeaders{}

	newCtx := InjectTraceIntoContext(context.Background(), emptyHeaders)

	// Should return the original context (no trace injection)
	span := oteltrace.SpanFromContext(newCtx)
	if span.SpanContext().HasSpanID() {
		t.Error("Expected no span ID for empty headers")
	}
}

func TestTracePropagationWithInvalidHeaders(t *testing.T) {
	// Test with invalid trace headers
	invalidHeaders := TraceHeaders{
		TraceID: "invalid-trace-id",
		SpanID:  "invalid-span-id",
	}

	newCtx := InjectTraceIntoContext(context.Background(), invalidHeaders)

	// Should return the original context (no trace injection)
	span := oteltrace.SpanFromContext(newCtx)
	if span.SpanContext().HasSpanID() {
		t.Error("Expected no span ID for invalid headers")
	}
}

func TestExtractTraceFromKafkaHeaders(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a test context with a span
	ctx, span := tp.Tracer("test").Start(context.Background(), "test.span")
	defer span.End()

	// Extract trace headers and serialize to Kafka headers
	traceHeaders := ExtractTraceFromContext(ctx)
	kafkaHeaders := SerializeTraceHeaders(traceHeaders)

	// Test extracting trace from Kafka headers
	newCtx := ExtractTraceFromKafkaHeaders(context.Background(), kafkaHeaders)

	// Check that trace context was injected
	newSpan := oteltrace.SpanFromContext(newCtx)
	if !newSpan.SpanContext().HasSpanID() {
		t.Error("Expected new context to have span ID")
	}

	// Verify trace IDs match
	if newSpan.SpanContext().TraceID() != span.SpanContext().TraceID() {
		t.Error("Expected trace IDs to match")
	}
}

func TestCreateChildSpan(t *testing.T) {
	// Set up a test trace provider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create a parent context with a span
	parentCtx, parentSpan := tp.Tracer("test").Start(context.Background(), "parent.span")
	defer parentSpan.End()

	// Create a child span
	childCtx, childSpan := CreateChildSpan(parentCtx, "child.span")
	defer childSpan.End()

	// Verify child span has a span ID
	if !childSpan.SpanContext().HasSpanID() {
		t.Error("Expected child span to have span ID")
	}

	// Verify trace IDs match (same trace)
	if childSpan.SpanContext().TraceID() != parentSpan.SpanContext().TraceID() {
		t.Error("Expected child and parent to have same trace ID")
	}

	// Verify span IDs are different
	if childSpan.SpanContext().SpanID() == parentSpan.SpanContext().SpanID() {
		t.Error("Expected child and parent to have different span IDs")
	}

	// Verify child context has the child span
	spanFromCtx := oteltrace.SpanFromContext(childCtx)
	if spanFromCtx.SpanContext().SpanID() != childSpan.SpanContext().SpanID() {
		t.Error("Expected context to contain child span")
	}
}

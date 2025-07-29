package pubsub_test

import (
	"context"
	"zensor-server/internal/infra/pubsub"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ = ginkgo.Describe("Trace Propagation", func() {
	ginkgo.Context("TracePropagation", func() {
		var (
			tp   *trace.TracerProvider
			ctx  context.Context
			span oteltrace.Span
		)

		ginkgo.BeforeEach(func() {
			// Set up a test trace provider
			tp = trace.NewTracerProvider(
				trace.WithSpanProcessor(tracetest.NewSpanRecorder()),
			)
			otel.SetTracerProvider(tp)

			// Create a test context with a span
			ctx, span = tp.Tracer("test").Start(context.Background(), "test.span")
		})

		ginkgo.AfterEach(func() {
			span.End()
			tp.Shutdown(context.Background())
		})

		ginkgo.It("should handle complete trace propagation cycle", func() {
			// Test extracting trace headers
			traceHeaders := pubsub.ExtractTraceFromContext(ctx)

			gomega.Expect(traceHeaders.TraceID).NotTo(gomega.BeEmpty())
			gomega.Expect(traceHeaders.SpanID).NotTo(gomega.BeEmpty())

			// Test serializing to Kafka headers
			kafkaHeaders := pubsub.SerializeTraceHeaders(traceHeaders)

			gomega.Expect(kafkaHeaders).NotTo(gomega.BeEmpty())
			gomega.Expect(string(kafkaHeaders["trace_id"])).To(gomega.Equal(traceHeaders.TraceID))
			gomega.Expect(string(kafkaHeaders["span_id"])).To(gomega.Equal(traceHeaders.SpanID))

			// Test deserializing from Kafka headers
			deserializedHeaders := pubsub.DeserializeTraceHeaders(kafkaHeaders)

			gomega.Expect(deserializedHeaders.TraceID).To(gomega.Equal(traceHeaders.TraceID))
			gomega.Expect(deserializedHeaders.SpanID).To(gomega.Equal(traceHeaders.SpanID))

			// Test injecting trace into context
			newCtx := pubsub.InjectTraceIntoContext(context.Background(), deserializedHeaders)

			// Check that trace context was injected
			newSpan := oteltrace.SpanFromContext(newCtx)
			gomega.Expect(newSpan.SpanContext().HasSpanID()).To(gomega.BeTrue())

			// Verify trace IDs match
			gomega.Expect(newSpan.SpanContext().TraceID()).To(gomega.Equal(span.SpanContext().TraceID()))
		})
	})

	ginkgo.Context("TracePropagationWithEmptyHeaders", func() {
		ginkgo.It("should handle empty headers gracefully", func() {
			// Test with empty headers
			emptyHeaders := pubsub.TraceHeaders{}

			newCtx := pubsub.InjectTraceIntoContext(context.Background(), emptyHeaders)

			// Should return the original context (no trace injection)
			span := oteltrace.SpanFromContext(newCtx)
			gomega.Expect(span.SpanContext().HasSpanID()).To(gomega.BeFalse())
		})
	})

	ginkgo.Context("TracePropagationWithInvalidHeaders", func() {
		ginkgo.It("should handle invalid headers gracefully", func() {
			// Test with invalid trace headers
			invalidHeaders := pubsub.TraceHeaders{
				TraceID: "invalid-trace-id",
				SpanID:  "invalid-span-id",
			}

			newCtx := pubsub.InjectTraceIntoContext(context.Background(), invalidHeaders)

			// Should return the original context (no trace injection)
			span := oteltrace.SpanFromContext(newCtx)
			gomega.Expect(span.SpanContext().HasSpanID()).To(gomega.BeFalse())
		})
	})

	ginkgo.Context("ExtractTraceFromKafkaHeaders", func() {
		ginkgo.It("should extract trace from valid Kafka headers", func() {
			// Create valid Kafka headers
			kafkaHeaders := map[string][]byte{
				"trace_id": []byte("1234567890abcdef1234567890abcdef"),
				"span_id":  []byte("1234567890abcdef"),
			}

			traceHeaders := pubsub.DeserializeTraceHeaders(kafkaHeaders)

			gomega.Expect(traceHeaders.TraceID).To(gomega.Equal("1234567890abcdef1234567890abcdef"))
			gomega.Expect(traceHeaders.SpanID).To(gomega.Equal("1234567890abcdef"))
		})

		ginkgo.It("should handle missing Kafka headers", func() {
			// Create headers with missing trace information
			kafkaHeaders := map[string][]byte{
				"other_header": []byte("value"),
			}

			traceHeaders := pubsub.DeserializeTraceHeaders(kafkaHeaders)

			gomega.Expect(traceHeaders.TraceID).To(gomega.BeEmpty())
			gomega.Expect(traceHeaders.SpanID).To(gomega.BeEmpty())
		})
	})

	ginkgo.Context("CreateChildSpan", func() {
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

		ginkgo.It("should create child span from context", func() {
			// Create a parent span
			parentCtx, parentSpan := tp.Tracer("test").Start(context.Background(), "parent.span")
			defer parentSpan.End()

			// Create a child span
			childCtx, childSpan := pubsub.CreateChildSpan(parentCtx, "child.span")
			defer childSpan.End()

			// Verify the child span is created
			gomega.Expect(childSpan).NotTo(gomega.BeNil())

			// Verify the child span has a parent
			childSpanContext := childSpan.SpanContext()
			parentSpanContext := parentSpan.SpanContext()

			gomega.Expect(childSpanContext.TraceID()).To(gomega.Equal(parentSpanContext.TraceID()))
			gomega.Expect(childSpanContext.SpanID()).NotTo(gomega.Equal(parentSpanContext.SpanID()))

			// Verify the context contains the child span
			spanFromCtx := oteltrace.SpanFromContext(childCtx)
			gomega.Expect(spanFromCtx.SpanContext().SpanID()).To(gomega.Equal(childSpanContext.SpanID()))
		})
	})
})

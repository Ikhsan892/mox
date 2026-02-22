package monitoring

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// setupTestTracer sets up a test trace provider with an in-memory exporter.
func setupTestTracer(t *testing.T) (*tracetest.InMemoryExporter, func()) {
	t.Helper()

	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exp),
	)
	otel.SetTracerProvider(tp)

	cleanup := func() {
		_ = tp.Shutdown(context.Background())
	}

	return exp, cleanup
}

func TestNewTraceContext(t *testing.T) {
	ctx := context.Background()
	tc := NewTraceContext(ctx)

	assert.NotNil(t, tc)
	assert.Equal(t, ctx, tc.otelCtx)
}

func TestTraceContext_WithTraceName(t *testing.T) {
	tc := NewTraceContext(context.Background())

	result := tc.WithTraceName("my-tracer")
	assert.Same(t, tc, result, "WithTraceName should return same instance for chaining")
	assert.Equal(t, "my-tracer", tc.traceName)
}

func TestTraceContext_WithSpanName(t *testing.T) {
	tc := NewTraceContext(context.Background())

	result := tc.WithSpanName("my-span")
	assert.Same(t, tc, result)
	assert.Equal(t, "my-span", tc.spanName)
}

func TestTraceContext_WithTraceparent(t *testing.T) {
	tc := NewTraceContext(context.Background())

	result := tc.WithTraceparent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	assert.Same(t, tc, result)
	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", tc.traceparent)
}

func TestBuild_NoTraceparent(t *testing.T) {
	exp, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx, span := NewTraceContext(context.Background()).
		WithTraceName("test-tracer").
		WithSpanName("new-span").
		Build()

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())
	assert.True(t, span.SpanContext().IsSampled())

	span.End()

	// Verify span was recorded
	spans := exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "new-span", spans[0].Name)
}

func TestBuild_WithValidTraceparent(t *testing.T) {
	_, cleanup := setupTestTracer(t)
	defer cleanup()

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	ctx, span := NewTraceContext(context.Background()).
		WithTraceName("test-tracer").
		WithSpanName("child-span").
		WithTraceparent(traceparent).
		Build()

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	// The span should be a child of the trace from traceparent
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", span.SpanContext().TraceID().String())

	span.End()
}

func TestBuild_WithInvalidTraceparent(t *testing.T) {
	_, cleanup := setupTestTracer(t)
	defer cleanup()

	// Invalid traceparent — should fallback to new trace (not crash)
	ctx, span := NewTraceContext(context.Background()).
		WithTraceName("test-tracer").
		WithSpanName("fallback-span").
		WithTraceparent("invalid-traceparent").
		Build()

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())

	span.End()
}

func TestBuild_Chaining(t *testing.T) {
	_, cleanup := setupTestTracer(t)
	defer cleanup()

	// Test the full builder chain
	ctx, span := NewTraceContext(context.Background()).
		WithTraceName("service-a").
		WithSpanName("operation-x").
		Build()

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	// Verify span context is embedded in the returned context
	spanFromCtx := trace.SpanFromContext(ctx)
	assert.Equal(t, span.SpanContext().TraceID(), spanFromCtx.SpanContext().TraceID())

	span.End()
}

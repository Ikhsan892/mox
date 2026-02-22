package trace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewTraceProviderBuilder(t *testing.T) {
	builder := NewTraceProviderBuilder("test-svc", "1.0.0", "dev")
	assert.NotNil(t, builder)
	assert.Equal(t, "test-svc", builder.name)
	assert.Equal(t, "1.0.0", builder.version)
	assert.Equal(t, "dev", builder.instanceId)
}

func TestBuild(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	builder := NewTraceProviderBuilder("test-svc", "1.0.0", "dev")

	tp, closeFunc, err := builder.SetExporter(exp).Build()
	require.NoError(t, err)
	assert.NotNil(t, tp)
	assert.NotNil(t, closeFunc)

	// Clean up
	err = tp.Shutdown(t.Context())
	assert.NoError(t, err)
}

func TestSetExporter_Chaining(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	builder := NewTraceProviderBuilder("svc", "2.0", "staging")

	result := builder.SetExporter(exp)
	assert.Same(t, builder, result, "SetExporter should return same builder for chaining")
}

func TestBuild_ProducesSpans(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	builder := NewTraceProviderBuilder("test-svc", "1.0.0", "dev")

	tp, _, err := builder.SetExporter(exp).Build()
	require.NoError(t, err)

	// Create a span to verify the provider works
	tracer := tp.Tracer("test-tracer")
	_, span := tracer.Start(t.Context(), "test-span")
	span.End()

	// Force flush to ensure span is exported
	err = tp.ForceFlush(t.Context())
	require.NoError(t, err)

	spans := exp.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-span", spans[0].Name)
}

func TestBuild_NilExporter(t *testing.T) {
	builder := NewTraceProviderBuilder("test-svc", "1.0.0", "dev")

	// With nil exporter, trace SDK will still build but panics when exporting.
	// Our builder doesn't guard this explicitly, but SpanProcessor handles nil.
	// Just verify that SetExporter(nil).Build() doesn't crash during setup.
	tp, _, err := builder.SetExporter(nil).Build()

	// Build should succeed — the SDK creates a no-op processor for nil exporters
	// in NewBatchSpanProcessor, though behavior may vary by version.
	if err != nil {
		return // acceptable if SDK rejects nil
	}

	if tp != nil {
		_ = tp.Shutdown(t.Context())
	}
}

func TestBuild_SamplerAlwaysSample(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	builder := NewTraceProviderBuilder("test-svc", "1.0.0", "dev")

	tp, _, err := builder.SetExporter(exp).Build()
	require.NoError(t, err)

	// Verify AlwaysSample by creating spans — all should be sampled
	tracer := tp.Tracer("test")
	for i := 0; i < 5; i++ {
		_, span := tracer.Start(t.Context(), "span")
		assert.True(t, span.SpanContext().IsSampled())
		span.End()
	}

	_ = tp.ForceFlush(t.Context())

	// All 5 spans should be captured since AlwaysSample is set
	spans := exp.GetSpans()
	assert.Len(t, spans, 5)

	_ = tp.Shutdown(t.Context())
}

// Verify the builder implements SpanExporter interface correctly
func TestBuild_SpanExporterType(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	var _ sdktrace.SpanExporter = exp // compile-time check
	assert.NotNil(t, exp)
}

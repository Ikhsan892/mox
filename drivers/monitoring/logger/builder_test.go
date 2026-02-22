package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// --- in-memory log exporter for testing ---

type memoryLogExporter struct{}

func newMemoryLogExporter() *memoryLogExporter {
	return &memoryLogExporter{}
}

func (m *memoryLogExporter) Export(_ context.Context, _ []sdklog.Record) error {
	return nil
}

func (m *memoryLogExporter) ForceFlush(_ context.Context) error {
	return nil
}

func (m *memoryLogExporter) Shutdown(_ context.Context) error {
	return nil
}

// --- tests ---

func TestNewLoggerProviderBuilder(t *testing.T) {
	builder := NewLoggerProviderBuilder("test-svc", "1.0.0", "dev")
	assert.NotNil(t, builder)
	assert.Equal(t, "test-svc", builder.name)
	assert.Equal(t, "1.0.0", builder.version)
	assert.Equal(t, "dev", builder.instanceId)
}

func TestLoggerBuild(t *testing.T) {
	exp := newMemoryLogExporter()
	builder := NewLoggerProviderBuilder("test-svc", "1.0.0", "dev")

	lp, closeFunc, err := builder.SetExporter(exp).Build()
	require.NoError(t, err)
	assert.NotNil(t, lp)
	assert.NotNil(t, closeFunc)

	// Clean up
	err = lp.Shutdown(t.Context())
	assert.NoError(t, err)
}

func TestLoggerSetExporter_Chaining(t *testing.T) {
	exp := newMemoryLogExporter()
	builder := NewLoggerProviderBuilder("svc", "2.0", "staging")

	result := builder.SetExporter(exp)
	assert.Same(t, builder, result, "SetExporter should return same builder for chaining")
}

// Verify the memoryLogExporter satisfies the Exporter interface used in the builder
func TestLoggerExporterType(t *testing.T) {
	exp := newMemoryLogExporter()
	var _ sdklog.Exporter = exp // compile-time check
	assert.NotNil(t, exp)
}

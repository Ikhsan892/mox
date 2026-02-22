package metric

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// --- in-memory exporter for testing ---

type memoryExporter struct{}

func newMemoryExporter() *memoryExporter {
	return &memoryExporter{}
}

func (m *memoryExporter) Temporality(_ metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (m *memoryExporter) Aggregation(_ metric.InstrumentKind) metric.Aggregation {
	return nil
}

func (m *memoryExporter) Export(_ context.Context, _ *metricdata.ResourceMetrics) error {
	return nil
}

func (m *memoryExporter) ForceFlush(_ context.Context) error {
	return nil
}

func (m *memoryExporter) Shutdown(_ context.Context) error {
	return nil
}

// --- tests ---

func TestNewMeterProviderBuilder(t *testing.T) {
	builder := NewMeterProviderBuilder("test-svc", "1.0.0", "dev")
	assert.NotNil(t, builder)
	assert.Equal(t, "test-svc", builder.name)
	assert.Equal(t, "1.0.0", builder.version)
	assert.Equal(t, "dev", builder.instanceId)
}

func TestBuild_NoExporter(t *testing.T) {
	builder := NewMeterProviderBuilder("test-svc", "1.0.0", "dev")

	mp, closeFunc, err := builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exporter is not set")
	assert.Nil(t, mp)
	assert.Nil(t, closeFunc)
}

func TestBuild_WithExporter(t *testing.T) {
	exp := newMemoryExporter()
	builder := NewMeterProviderBuilder("test-svc", "1.0.0", "dev")

	mp, closeFunc, err := builder.SetExporter(exp).Build()
	require.NoError(t, err)
	assert.NotNil(t, mp)
	assert.NotNil(t, closeFunc)

	// Clean up
	err = mp.Shutdown(t.Context())
	assert.NoError(t, err)
}

func TestSetExporter_Chaining(t *testing.T) {
	exp := newMemoryExporter()
	builder := NewMeterProviderBuilder("svc", "2.0", "staging")

	result := builder.SetExporter(exp)
	assert.Same(t, builder, result, "SetExporter should return same builder for chaining")
}

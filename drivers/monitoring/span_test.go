package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestConstructNewSpanContext(t *testing.T) {
	t.Parallel()

	req := NewRequest{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
	}

	sc, err := ConstructNewSpanContext(req)
	require.NoError(t, err)
	assert.True(t, sc.IsValid())
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", sc.TraceID().String())
	assert.Equal(t, "00f067aa0ba902b7", sc.SpanID().String())
	assert.False(t, sc.IsRemote())
}

func TestConstructNewSpanContext_InvalidTraceID(t *testing.T) {
	t.Parallel()

	req := NewRequest{
		TraceID: "not-a-valid-hex",
		SpanID:  "00f067aa0ba902b7",
	}

	_, err := ConstructNewSpanContext(req)
	assert.Error(t, err)
}

func TestConstructNewSpanContext_InvalidSpanID(t *testing.T) {
	t.Parallel()

	req := NewRequest{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "invalid",
	}

	_, err := ConstructNewSpanContext(req)
	assert.Error(t, err)
}

func TestConstructNewSpanContextFromTraceparent(t *testing.T) {
	t.Parallel()

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	sc, err := ConstructNewSpanContextFromTraceparent(traceparent)
	require.NoError(t, err)
	assert.True(t, sc.IsValid())
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", sc.TraceID().String())
	assert.Equal(t, "00f067aa0ba902b7", sc.SpanID().String())
	assert.True(t, sc.IsSampled())
	assert.True(t, sc.IsRemote())
}

func TestConstructNewSpanContextFromTraceparent_NotSampled(t *testing.T) {
	t.Parallel()

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"

	sc, err := ConstructNewSpanContextFromTraceparent(traceparent)
	require.NoError(t, err)
	assert.True(t, sc.IsValid())
	assert.False(t, sc.IsSampled())
}

func TestConstructNewSpanContextFromTraceparent_InvalidFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		traceparent string
	}{
		{"empty string", ""},
		{"too few parts", "00-4bf92f3577b34da6a3ce929d0e0e4736"},
		{"too many parts", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra"},
		{"no dashes", "abcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ConstructNewSpanContextFromTraceparent(tc.traceparent)
			assert.Error(t, err)
		})
	}
}

func TestConstructNewSpanContextFromTraceparent_InvalidTraceID(t *testing.T) {
	t.Parallel()

	traceparent := "00-invalid_trace_id_here_padding_pad-00f067aa0ba902b7-01"

	_, err := ConstructNewSpanContextFromTraceparent(traceparent)
	assert.Error(t, err)
}

func TestConstructNewSpanContextFromTraceparent_InvalidSpanID(t *testing.T) {
	t.Parallel()

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-invalid_span_id-01"

	_, err := ConstructNewSpanContextFromTraceparent(traceparent)
	assert.Error(t, err)
}

func TestNewRequest_ZeroValue(t *testing.T) {
	t.Parallel()

	req := NewRequest{}
	assert.Equal(t, "", req.TraceID)
	assert.Equal(t, "", req.SpanID)
	assert.Equal(t, "", req.Requestid)
}

// Verify trace.SpanContext interface methods
func TestSpanContext_Interface(t *testing.T) {
	t.Parallel()

	traceparent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	sc, err := ConstructNewSpanContextFromTraceparent(traceparent)
	require.NoError(t, err)

	// Check HasTraceID and HasSpanID
	assert.True(t, sc.HasTraceID())
	assert.True(t, sc.HasSpanID())

	// Verify TraceFlags
	assert.Equal(t, trace.FlagsSampled, sc.TraceFlags())
}

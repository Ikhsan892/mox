package monitoring

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// TraceContextRequest encapsulates the OpenTelemetry trace context
type TraceContextRequest struct {
	traceName   string
	spanName    string
	otelCtx     context.Context
	spanCtx     trace.SpanContext
	traceparent string
}

// NewTraceContext initializes a new instance of TraceContextRequest
func NewTraceContext(ctx context.Context) *TraceContextRequest {
	return &TraceContextRequest{
		otelCtx: ctx,
	}
}

// WithTraceName sets the trace name
func (t *TraceContextRequest) WithTraceName(traceName string) *TraceContextRequest {
	t.traceName = traceName
	return t
}

// WithSpanName sets the span name
func (t *TraceContextRequest) WithSpanName(spanName string) *TraceContextRequest {
	t.spanName = spanName
	return t
}

// WithTraceparent sets the traceparent header
func (t *TraceContextRequest) WithTraceparent(traceparent string) *TraceContextRequest {
	t.traceparent = traceparent
	return t
}

// Build constructs the final tracing context and span
func (t *TraceContextRequest) Build() (context.Context, trace.Span) {
	tracer := otel.Tracer(t.traceName)

	// If no traceparent is provided, start a new trace
	if t.traceparent == "" {
		ctx, span := tracer.Start(t.otelCtx, t.spanName)
		return ctx, span
	}

	// Parse traceparent to get span context
	spanContext, err := ConstructNewSpanContextFromTraceparent(t.traceparent)
	if err != nil {
		fmt.Println("Error parsing traceparent:", err)
		ctx, span := tracer.Start(t.otelCtx, t.spanName)
		return ctx, span
	}

	// Inject existing trace context and start a new span
	t.otelCtx = trace.ContextWithSpanContext(t.otelCtx, spanContext)
	ctx, span := tracer.Start(t.otelCtx, t.spanName)
	return ctx, span
}

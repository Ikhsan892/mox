package monitoring

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type NewRequest struct {
	Requestid string `json: "requestid"`
	TraceID   string
	SpanID    string
}

func ConstructNewSpanContext(request NewRequest) (spanContext trace.SpanContext, err error) {
	var traceID trace.TraceID
	traceID, err = trace.TraceIDFromHex(request.TraceID)
	if err != nil {
		fmt.Println("error: ", err)
		return spanContext, err
	}
	var spanID trace.SpanID
	spanID, err = trace.SpanIDFromHex(request.SpanID)
	if err != nil {
		fmt.Println("error: ", err)
		return spanContext, err
	}
	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = false
	spanContext = trace.NewSpanContext(spanContextConfig)
	return spanContext, nil
}

func ConstructNewSpanContextFromTraceparent(traceparent string) (trace.SpanContext, error) {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return trace.SpanContext{}, fmt.Errorf("invalid traceparent format")
	}

	traceID, err := trace.TraceIDFromHex(parts[1])
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("invalid trace ID: %w", err)
	}

	spanID, err := trace.SpanIDFromHex(parts[2])
	if err != nil {
		return trace.SpanContext{}, fmt.Errorf("invalid span ID: %w", err)
	}

	traceFlags := trace.TraceFlags(0)
	if parts[3] == "01" {
		traceFlags = trace.FlagsSampled
	}

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
		Remote:     true, // Since it came from KrakenD
	})

	return spanContext, nil
}

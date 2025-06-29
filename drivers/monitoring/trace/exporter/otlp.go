package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
)

func NewOTLP(endpoint string) (*otlptrace.Exporter, error) {
	ctx := context.Background()
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return nil, err
	}
	return traceExp, nil
}

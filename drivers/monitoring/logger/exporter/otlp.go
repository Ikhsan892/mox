package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
)

func NewOTLP(endpoint string) (log.Exporter, error) {
	ctx := context.Background()

	logClient, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithInsecure(),

		otlploggrpc.WithDialOption(grpc.WithBlock()),
		otlploggrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}

	return logClient, nil
}

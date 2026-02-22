package metric

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CloseFunc func(ctx context.Context) error

func NewMeterProviderBuilder(name, version, instanceId string) *meterProviderBuilder {
	return &meterProviderBuilder{
		name:       name,
		version:    version,
		instanceId: instanceId,
	}
}

type meterProviderBuilder struct {
	name       string
	version    string
	instanceId string
	exporter   metric.Exporter
}

func (b *meterProviderBuilder) SetExporter(exp metric.Exporter) *meterProviderBuilder {
	b.exporter = exp
	return b
}

func (b *meterProviderBuilder) Build() (*metric.MeterProvider, CloseFunc, error) {
	if b.exporter == nil {
		return nil, nil, fmt.Errorf("exporter is not set")
	}

	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(b.name),
			semconv.ServiceVersionKey.String(b.version),
			semconv.ServiceInstanceIDKey.String(b.instanceId),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(
				b.exporter,
				metric.WithInterval(3*time.Second),
			),
		),
	)

	return meterProvider, func(ctx context.Context) error {
		cxt, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.exporter.Shutdown(cxt); err != nil {
			return err
		}
		return nil
	}, nil
}

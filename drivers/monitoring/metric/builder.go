package metric

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type CloseFunc func(ctx context.Context) error

func NewMeterProviderBuilder() *meterProviderBuilder {
	return &meterProviderBuilder{}
}

type meterProviderBuilder struct {
	exporter metric.Exporter
	res      *resource.Resource
}

func (b *meterProviderBuilder) SetResource(res *resource.Resource) *meterProviderBuilder {
	b.res = res
	return b
}

func (b *meterProviderBuilder) SetExporter(exp metric.Exporter) *meterProviderBuilder {
	b.exporter = exp
	return b
}

func (b meterProviderBuilder) Build() (*metric.MeterProvider, error) {
	if b.exporter == nil {
		return nil, fmt.Errorf("exporter is not set")
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(b.res),
		metric.WithReader(
			metric.NewPeriodicReader(
				b.exporter,
				metric.WithInterval(3*time.Second),
			),
		),
	)
	return meterProvider, nil
}

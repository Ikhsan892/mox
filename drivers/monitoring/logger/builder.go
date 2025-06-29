package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type CloseFunc func(ctx context.Context) error

func NewLoggerProviderBuilder(name, version, instanceId string) *loggerProviderBuilder {
	return &loggerProviderBuilder{
		name:       name,
		version:    version,
		instanceId: instanceId,
	}
}

type loggerProviderBuilder struct {
	name       string
	version    string
	instanceId string
	exporter   log.Exporter
}

func (b *loggerProviderBuilder) SetExporter(exp log.Exporter) *loggerProviderBuilder {
	b.exporter = exp
	return b
}

func (b *loggerProviderBuilder) Build() (*log.LoggerProvider, CloseFunc, error) {
	ctx := context.Background()
	res, err := resource.New(ctx,
		// resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// the service name used to display traces in backend
			semconv.ServiceNameKey.String(b.name),
			semconv.ServiceVersionKey.String(b.version),
			semconv.ServiceInstanceIDKey.String(b.instanceId),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(
			log.NewBatchProcessor(b.exporter, log.WithExportInterval(3*time.Second)),
		),
	)

	return loggerProvider, func(ctx context.Context) error {
		cxt, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.exporter.Shutdown(cxt); err != nil {
			return err
		}
		return err
	}, nil
}

package monitoring

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"mox/pkg/driver"

	tlogger "mox/drivers/monitoring/logger"
	logExporter "mox/drivers/monitoring/logger/exporter"
	tmetric "mox/drivers/monitoring/metric"
	metricExporter "mox/drivers/monitoring/metric/exporter"
	ttrace "mox/drivers/monitoring/trace"
	traceExporter "mox/drivers/monitoring/trace/exporter"
	core "mox/internal"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	OTEL_DRIVER                  = "OTEL"
	_           (driver.IDriver) = (*Otel)(nil)
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(app core.App, ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(app)
	if err != nil {
		app.Logger().Error(err.Error())
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	if app.Config().Monitoring.EnableTelemetry {
		meterProvider, err := newMeterProvider(app)
		if err != nil {
			app.Logger().Error(err.Error())
			handleErr(err)
			return nil, err
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	} else {
		app.Logger().Info("metric collection is disabled")
	}

	// Set up log provider
	if app.Config().Monitoring.EnableCollectLog {
		loggerProvider, err := newLoggerProvider(app)
		if err != nil {
			app.Logger().Error(err.Error())
			handleErr(err)
			return nil, err
		}
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	} else {
		app.Logger().Info("collecting log is disabled")
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(app core.App) (*trace.TracerProvider, error) {
	spanExporter, err := traceExporter.NewOTLP(app.Config().Monitoring.OtelEndpoint)
	if err != nil {
		return nil, err
	}

	app.Logger().Info("[TRACER] OTLP Connected")

	cfg := app.Config()

	tracerProvider, _, err := ttrace.NewTraceProviderBuilder(
		cfg.App.Name,
		strconv.Itoa(cfg.App.Version),
		cfg.App.Mode,
	).SetExporter(spanExporter).Build()
	if err != nil {
		return nil, err
	}

	app.Logger().Debug("Tracer set to auth-service")

	return tracerProvider, nil
}

func newMeterProvider(app core.App) (*metric.MeterProvider, error) {
	mExporter, err := metricExporter.NewOTLP(app.Config().Monitoring.OtelEndpoint)
	if err != nil {
		return nil, err
	}

	app.Logger().Info("[METRIC] OTLP Connected")

	cfg := app.Config()

	meterProvider, _, err := tmetric.NewMeterProviderBuilder(
		cfg.App.Name,
		strconv.Itoa(cfg.App.Version),
		cfg.App.Mode,
	).SetExporter(mExporter).Build()
	if err != nil {
		return nil, err
	}

	app.Logger().Debug("Meter provider initialized")

	return meterProvider, nil
}

func newLoggerProvider(app core.App) (*log.LoggerProvider, error) {
	spanExporter, err := logExporter.NewOTLP(app.Config().Monitoring.OtelEndpoint)
	if err != nil {
		return nil, err
	}

	app.Logger().Info("[LOGGER] OTLP Connected")

	cfg := app.Config()

	loggerProvider, _, err := tlogger.NewLoggerProviderBuilder(
		cfg.App.Name,
		strconv.Itoa(cfg.App.Version),
		cfg.App.Mode,
	).SetExporter(spanExporter).Build()
	if err != nil {
		return nil, err
	}

	app.Logger().Debug("logger set to auth-service")

	return loggerProvider, nil
}

type Otel struct {
	app      core.App
	ctx      context.Context
	shutdown func(context.Context) error
	mu       *sync.Mutex
}

func NewOtel(app core.App) *Otel {
	return &Otel{app: app, mu: &sync.Mutex{}, ctx: context.Background()}
}

// Close implements driver.IDriver.
func (o *Otel) Close() error {
	return o.shutdown(o.ctx)
}

// Init implements driver.IDriver.
func (o *Otel) Init() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.app.Logger().Info("Initializing Otel")

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(o.app, o.ctx)
	if err != nil {
		return err
	}

	o.app.Logger().Info("Otel is initialized")

	o.shutdown = otelShutdown

	return nil
}

// Instance implements driver.IDriver.
func (o *Otel) Instance() interface{} {
	return nil
}

// Name implements driver.IDriver.
func (o *Otel) Name() string {
	return OTEL_DRIVER
}

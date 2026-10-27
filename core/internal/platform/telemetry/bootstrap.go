package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func New(
	ctx context.Context,
	settings config.Telemetry,
	info buildinfo.Info,
	logger *slog.Logger,
) (*Runtime, error) {
	if ctx == nil {
		return nil, fmt.Errorf("telemetry bootstrap context is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("telemetry structured logger is required")
	}
	instanceID, err := randomid.NewUUIDv4()
	if err != nil {
		return nil, fmt.Errorf("telemetry service instance ID could not be generated")
	}
	res, err := newResource(info, instanceID)
	if err != nil {
		return nil, err
	}
	failures := newFailureReporter(logger)
	if settings.Enabled() {
		return newOTLPRuntime(ctx, settings, res, failures)
	}
	tracerProvider := tracenoop.NewTracerProvider()
	meterProvider := metricnoop.NewMeterProvider()
	return assembleRuntime(
		tracerProvider,
		meterProvider,
		res,
		failures,
		func(context.Context) error { return nil },
		func(context.Context) error { return nil },
	), nil
}

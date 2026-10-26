package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	spanBatchTimeout = 5 * time.Second
	spanQueueSize    = 2048
	spanBatchSize    = 512
	metricInterval   = 30 * time.Second
)

func newOTLPRuntime(
	ctx context.Context,
	settings config.Telemetry,
	res *resource.Resource,
	failures *failureReporter,
) (*Runtime, error) {
	spanExporter, metricExporter, err := newOTLPExporters(ctx, settings, failures)
	if err != nil {
		return nil, err
	}
	return newProviderRuntime(ctx, settings, res, failures, spanExporter, metricExporter)
}

func newProviderRuntime(
	ctx context.Context,
	settings config.Telemetry,
	res *resource.Resource,
	failures *failureReporter,
	spanExporter sdktrace.SpanExporter,
	metricExporter sdkmetric.Exporter,
) (*Runtime, error) {
	if spanExporter == nil || metricExporter == nil {
		return nil, fmt.Errorf("telemetry exporters are required")
	}
	spanProcessor := sdktrace.NewBatchSpanProcessor(
		observedSpanExporter{delegate: spanExporter, failures: failures},
		sdktrace.WithMaxQueueSize(spanQueueSize),
		sdktrace.WithMaxExportBatchSize(spanBatchSize),
		sdktrace.WithBatchTimeout(spanBatchTimeout),
		sdktrace.WithExportTimeout(settings.ExportTimeout()),
	)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(settings.TraceSampleRatio()),
		)),
		sdktrace.WithSpanProcessor(spanProcessor),
	)
	reader := sdkmetric.NewPeriodicReader(
		observedMetricExporter{delegate: metricExporter, failures: failures},
		sdkmetric.WithInterval(metricInterval),
		sdkmetric.WithTimeout(settings.ExportTimeout()),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	if err := failures.bind(meterProvider); err != nil {
		cleanupProviders(ctx, tracerProvider, meterProvider, failures)
		return nil, fmt.Errorf("telemetry providers could not be initialized")
	}
	return assembleRuntime(
		tracerProvider,
		meterProvider,
		res,
		failures,
		tracerProvider.Shutdown,
		meterProvider.Shutdown,
	), nil
}

func cleanupProviders(
	ctx context.Context,
	tracerProvider *sdktrace.TracerProvider,
	meterProvider *sdkmetric.MeterProvider,
	failures *failureReporter,
) {
	if err := tracerProvider.Shutdown(ctx); err != nil {
		failures.report(ctx, signalTrace, operationShutdown)
	}
	if err := meterProvider.Shutdown(ctx); err != nil {
		failures.report(ctx, signalMetric, operationShutdown)
	}
}

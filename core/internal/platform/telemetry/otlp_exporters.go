package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"maps"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	retryInitialInterval = 500 * time.Millisecond
	retryMaximumInterval = 2 * time.Second
)

func newOTLPExporters(
	ctx context.Context,
	settings config.Telemetry,
	failures *failureReporter,
) (sdktrace.SpanExporter, sdkmetric.Exporter, error) {
	headers, err := exportHeaders(settings)
	if err != nil {
		return nil, nil, err
	}
	traceExporter, err := otlptracehttp.New(ctx, traceExporterOptions(settings, headers)...)
	if err != nil {
		failures.report(ctx, signalTrace, operationInitialize)
		return nil, nil, fmt.Errorf("OTLP trace exporter could not be initialized")
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricExporterOptions(settings, headers)...)
	if err != nil {
		failures.report(ctx, signalMetric, operationInitialize)
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), settings.ExportTimeout())
		defer cancel()
		if shutdownErr := traceExporter.Shutdown(cleanupCtx); shutdownErr != nil {
			failures.report(cleanupCtx, signalTrace, operationShutdown)
		}
		return nil, nil, fmt.Errorf("OTLP metric exporter could not be initialized")
	}
	return traceExporter, metricExporter, nil
}

func exportHeaders(settings config.Telemetry) (map[string]string, error) {
	secret, configured := settings.Headers()
	if !configured {
		return nil, nil
	}
	return parseExportHeaders(secret.Value())
}

func traceExporterOptions(
	settings config.Telemetry,
	headers map[string]string,
) []otlptracehttp.Option {
	timeout := settings.ExportTimeout()
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(settings.Endpoint()),
		otlptracehttp.WithHeaders(maps.Clone(headers)),
		otlptracehttp.WithTimeout(timeout),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled: true, InitialInterval: retryInitialInterval,
			MaxInterval: retryMaximumInterval, MaxElapsedTime: timeout,
		}),
	}
	if settings.Insecure() {
		return append(options, otlptracehttp.WithInsecure())
	}
	return append(options, otlptracehttp.WithTLSClientConfig(&tls.Config{MinVersion: tls.VersionTLS12}))
}

func metricExporterOptions(
	settings config.Telemetry,
	headers map[string]string,
) []otlpmetrichttp.Option {
	timeout := settings.ExportTimeout()
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(settings.Endpoint()),
		otlpmetrichttp.WithHeaders(maps.Clone(headers)),
		otlpmetrichttp.WithTimeout(timeout),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithTemporalitySelector(sdkmetric.DefaultTemporalitySelector),
		otlpmetrichttp.WithAggregationSelector(sdkmetric.DefaultAggregationSelector),
		otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
			Enabled: true, InitialInterval: retryInitialInterval,
			MaxInterval: retryMaximumInterval, MaxElapsedTime: timeout,
		}),
	}
	if settings.Insecure() {
		return append(options, otlpmetrichttp.WithInsecure())
	}
	return append(options, otlpmetrichttp.WithTLSClientConfig(&tls.Config{MinVersion: tls.VersionTLS12}))
}

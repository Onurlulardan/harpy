package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	signalMetric        = "metric"
	signalTrace         = "trace"
	operationExport     = "export"
	operationFlush      = "force_flush"
	operationInitialize = "initialize"
	operationShutdown   = "shutdown"
)

type failureReporter struct {
	logger         *slog.Logger
	traceFailures  atomic.Uint64
	metricFailures atomic.Uint64
	counterMu      sync.RWMutex
	counter        metric.Int64Counter
}

func newFailureReporter(logger *slog.Logger) *failureReporter {
	return &failureReporter{logger: logger}
}

func (reporter *failureReporter) bind(provider metric.MeterProvider) error {
	counter, err := provider.Meter(inboundInstrumentation).Int64Counter(
		"harpy.telemetry.export.failures",
		metric.WithDescription("Telemetry export and flush failures."),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return fmt.Errorf("telemetry failure counter could not be created")
	}
	reporter.counterMu.Lock()
	reporter.counter = counter
	reporter.counterMu.Unlock()
	return nil
}

func (reporter *failureReporter) report(ctx context.Context, signal, operation string) {
	if ctx == nil {
		ctx = context.Background()
	}
	switch signal {
	case signalTrace:
		reporter.traceFailures.Add(1)
	case signalMetric:
		reporter.metricFailures.Add(1)
	}
	reporter.logger.ErrorContext(ctx, "telemetry delivery failed",
		"telemetry.signal", signal,
		"telemetry.operation", operation,
	)
	reporter.counterMu.RLock()
	counter := reporter.counter
	reporter.counterMu.RUnlock()
	if counter != nil {
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("telemetry.signal", signal),
			attribute.String("telemetry.operation", operation),
		))
	}
}

func (reporter *failureReporter) snapshot() FailureCounts {
	return FailureCounts{
		Trace:  reporter.traceFailures.Load(),
		Metric: reporter.metricFailures.Load(),
	}
}

type observedSpanExporter struct {
	delegate sdktrace.SpanExporter
	failures *failureReporter
}

func (exporter observedSpanExporter) ExportSpans(
	ctx context.Context,
	spans []sdktrace.ReadOnlySpan,
) error {
	if err := exporter.delegate.ExportSpans(ctx, spans); err != nil {
		exporter.failures.report(ctx, signalTrace, operationExport)
	}
	return nil
}

func (exporter observedSpanExporter) Shutdown(ctx context.Context) error {
	if err := exporter.delegate.Shutdown(ctx); err != nil {
		return fmt.Errorf("trace exporter shutdown failed")
	}
	return nil
}

type observedMetricExporter struct {
	delegate sdkmetric.Exporter
	failures *failureReporter
}

func (exporter observedMetricExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return exporter.delegate.Temporality(kind)
}

func (exporter observedMetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return exporter.delegate.Aggregation(kind)
}

func (exporter observedMetricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	if err := exporter.delegate.Export(ctx, data); err != nil {
		exporter.failures.report(ctx, signalMetric, operationExport)
	}
	return nil
}

func (exporter observedMetricExporter) ForceFlush(ctx context.Context) error {
	if err := exporter.delegate.ForceFlush(ctx); err != nil {
		exporter.failures.report(ctx, signalMetric, operationFlush)
		return fmt.Errorf("metric exporter force flush failed")
	}
	return nil
}

func (exporter observedMetricExporter) Shutdown(ctx context.Context) error {
	if err := exporter.delegate.Shutdown(ctx); err != nil {
		return fmt.Errorf("metric exporter shutdown failed")
	}
	return nil
}

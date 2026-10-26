package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type asyncRetentionInstruments struct {
	cycles        metric.Int64Counter
	duration      metric.Float64Histogram
	rows          metric.Int64Counter
	failures      metric.Int64Counter
	backlogYields metric.Int64Counter
}

func newAsyncRetentionInstruments(
	provider metric.MeterProvider,
) (asyncRetentionInstruments, error) {
	if provider == nil {
		return asyncRetentionInstruments{}, fmt.Errorf("async retention meter is required")
	}
	meter := provider.Meter(asyncInstrumentation)
	cycles, err := meter.Int64Counter(
		"harpy.async.retention.cycles",
		metric.WithDescription("Completed async retention cycles by fixed outcome."),
		metric.WithUnit("{cycle}"),
	)
	if err != nil {
		return asyncRetentionInstruments{}, fmt.Errorf("create retention cycle counter: %w", err)
	}
	duration, err := meter.Float64Histogram(
		"harpy.async.retention.duration",
		metric.WithDescription("Async retention cycle duration by fixed outcome."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return asyncRetentionInstruments{}, fmt.Errorf("create retention duration: %w", err)
	}
	rows, err := meter.Int64Counter(
		"harpy.async.retention.rows",
		metric.WithDescription("Rows changed by each fixed retention step."),
		metric.WithUnit("{row}"),
	)
	if err != nil {
		return asyncRetentionInstruments{}, fmt.Errorf("create retention row counter: %w", err)
	}
	failures, err := meter.Int64Counter(
		"harpy.async.retention.failures",
		metric.WithDescription("Failed async retention cycles."),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return asyncRetentionInstruments{}, fmt.Errorf("create retention failure counter: %w", err)
	}
	yields, err := meter.Int64Counter(
		"harpy.async.retention.backlog_yields",
		metric.WithDescription("Yields after four consecutive productive cycles."),
		metric.WithUnit("{yield}"),
	)
	if err != nil {
		return asyncRetentionInstruments{}, fmt.Errorf("create retention yield counter: %w", err)
	}
	return asyncRetentionInstruments{
		cycles: cycles, duration: duration, rows: rows,
		failures: failures, backlogYields: yields,
	}, nil
}

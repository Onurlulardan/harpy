package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

const asyncInstrumentation = "github.com/Onurlulardan/harpy/platform/async"

type asyncBacklogInstruments struct {
	oldestUndeliveredOutboxAge metric.Int64ObservableGauge
	undeliveredOutboxAlarm     metric.Int64ObservableGauge
	oldestDueJobLag            metric.Int64ObservableGauge
	parkedJobs                 metric.Int64ObservableGauge
	parkedOutbox               metric.Int64ObservableGauge
	failures                   metric.Int64ObservableCounter
}

func newAsyncBacklogInstruments(provider metric.MeterProvider) (asyncBacklogInstruments, error) {
	if provider == nil {
		return asyncBacklogInstruments{}, fmt.Errorf("async meter provider is required")
	}
	meter := provider.Meter(asyncInstrumentation)
	outboxAge, err := meter.Int64ObservableGauge(
		"harpy.async.outbox.undelivered.oldest_age",
		metric.WithDescription("Age of the oldest undelivered outbox event."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async outbox age gauge: %w", err)
	}
	outboxAlarm, err := meter.Int64ObservableGauge(
		"harpy.async.outbox.undelivered.alarm",
		metric.WithDescription("One when undelivered outbox age reaches 300 seconds."),
		metric.WithUnit("1"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async outbox alarm gauge: %w", err)
	}
	jobLag, err := meter.Int64ObservableGauge(
		"harpy.async.job.due.oldest_lag",
		metric.WithDescription("Lag of the oldest due, unparked job."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async due job lag gauge: %w", err)
	}
	parkedJobs, err := meter.Int64ObservableGauge(
		"harpy.async.job.parked",
		metric.WithDescription("Current parked job count."),
		metric.WithUnit("{job}"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async parked job gauge: %w", err)
	}
	parkedOutbox, err := meter.Int64ObservableGauge(
		"harpy.async.outbox.parked",
		metric.WithDescription("Current parked outbox event count."),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async parked outbox gauge: %w", err)
	}
	failures, err := meter.Int64ObservableCounter(
		"harpy.async.observer.failures",
		metric.WithDescription("Async backlog observation failures."),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return asyncBacklogInstruments{}, fmt.Errorf("create async observer failure counter: %w", err)
	}
	return asyncBacklogInstruments{
		oldestUndeliveredOutboxAge: outboxAge,
		undeliveredOutboxAlarm:     outboxAlarm,
		oldestDueJobLag:            jobLag, parkedJobs: parkedJobs,
		parkedOutbox: parkedOutbox, failures: failures,
	}, nil
}

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	asyncobservability "github.com/Onurlulardan/harpy/core/internal/platform/async/observability"
	"go.opentelemetry.io/otel/metric"
)

const asyncObservationFailureCode = "ASYNC_BACKLOG_OBSERVATION_FAILED"

type asyncBacklogObserver struct {
	source      asyncobservability.Source
	logger      *slog.Logger
	instruments asyncBacklogInstruments
	failures    atomic.Int64
}

func (runtime *Runtime) RegisterAsyncBacklog(
	source asyncobservability.Source,
	logger *slog.Logger,
) error {
	if runtime == nil {
		return fmt.Errorf("telemetry runtime is required")
	}
	return registerAsyncBacklog(source, runtime.meterProvider, logger)
}

func registerAsyncBacklog(
	source asyncobservability.Source,
	provider metric.MeterProvider,
	logger *slog.Logger,
) error {
	if source == nil || logger == nil {
		return fmt.Errorf("async observability dependencies are incomplete")
	}
	instruments, err := newAsyncBacklogInstruments(provider)
	if err != nil {
		return err
	}
	observer := &asyncBacklogObserver{
		source: source, logger: logger, instruments: instruments,
	}
	_, err = provider.Meter(asyncInstrumentation).RegisterCallback(
		observer.collect,
		instruments.oldestUndeliveredOutboxAge,
		instruments.undeliveredOutboxAlarm,
		instruments.oldestDueJobLag,
		instruments.parkedJobs,
		instruments.parkedOutbox,
		instruments.failures,
	)
	if err != nil {
		return fmt.Errorf("register async backlog observer: %w", err)
	}
	return nil
}

func (observer *asyncBacklogObserver) collect(
	ctx context.Context,
	sink metric.Observer,
) error {
	snapshot, err := observer.source.Snapshot(ctx)
	if err != nil {
		failures := observer.failures.Add(1)
		sink.ObserveInt64(observer.instruments.failures, failures)
		observer.logger.ErrorContext(ctx, "async backlog observation failed",
			"telemetry.component", "platform_async",
			"telemetry.operation", "observe_backlog",
			"failure.code", asyncObservationFailureCode,
		)
		return nil
	}
	sink.ObserveInt64(observer.instruments.oldestUndeliveredOutboxAge,
		snapshot.OldestUndeliveredOutboxAgeSeconds())
	sink.ObserveInt64(observer.instruments.undeliveredOutboxAlarm,
		snapshot.UndeliveredOutboxAlarm())
	sink.ObserveInt64(observer.instruments.oldestDueJobLag,
		snapshot.OldestDueJobLagSeconds())
	sink.ObserveInt64(observer.instruments.parkedJobs, snapshot.ParkedJobs())
	sink.ObserveInt64(observer.instruments.parkedOutbox, snapshot.ParkedOutbox())
	sink.ObserveInt64(observer.instruments.failures, observer.failures.Load())
	return nil
}

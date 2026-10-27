package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const retentionFailureOutcome = "failure"

type asyncRetentionObserver struct{ instruments asyncRetentionInstruments }

type retentionStepValue struct {
	name  string
	value int32
}

func (runtime *Runtime) NewAsyncRetentionObserver() (retention.Observer, error) {
	if runtime == nil {
		return nil, fmt.Errorf("telemetry runtime is required")
	}
	return newAsyncRetentionObserver(runtime.meterProvider)
}

func newAsyncRetentionObserver(
	provider metric.MeterProvider,
) (retention.Observer, error) {
	instruments, err := newAsyncRetentionInstruments(provider)
	if err != nil {
		return nil, err
	}
	return &asyncRetentionObserver{instruments: instruments}, nil
}

func (observer *asyncRetentionObserver) ObserveCycle(
	ctx context.Context,
	value retention.CycleObservation,
) {
	outcome := attribute.String("outcome", string(value.Outcome()))
	observer.instruments.cycles.Add(ctx, 1, metric.WithAttributes(outcome))
	observer.instruments.duration.Record(
		ctx, value.Duration().Seconds(), metric.WithAttributes(outcome),
	)
	counts := value.Counts()
	steps := [...]retentionStepValue{
		{"jobs_redacted", counts.JobsRedacted},
		{"outbox_redacted", counts.OutboxRedacted},
		{"dead_letters_redacted", counts.DeadLettersRedacted},
		{"inbox_purged", counts.InboxPurged},
		{"idempotency_purged", counts.IdempotencyPurged},
		{"jobs_purged", counts.JobsPurged},
		{"dead_letters_purged", counts.DeadLettersPurged},
	}
	for _, step := range steps {
		observer.instruments.rows.Add(ctx, int64(step.value), metric.WithAttributes(
			attribute.String("step", step.name),
		))
	}
}

func (observer *asyncRetentionObserver) ObserveFailure(
	ctx context.Context,
	duration time.Duration,
) {
	outcome := attribute.String("outcome", retentionFailureOutcome)
	observer.instruments.cycles.Add(ctx, 1, metric.WithAttributes(outcome))
	observer.instruments.duration.Record(
		ctx, duration.Seconds(), metric.WithAttributes(outcome),
	)
	observer.instruments.failures.Add(ctx, 1)
}

func (observer *asyncRetentionObserver) ObserveBacklogYield(ctx context.Context) {
	observer.instruments.backlogYields.Add(ctx, 1)
}

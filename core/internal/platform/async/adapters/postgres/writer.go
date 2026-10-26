package postgres

import (
	"context"
	"errors"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/idempotency"
)

var (
	ErrAuthorityMismatch   = errors.New("async input differs from bound owner authority")
	ErrIdempotencyConflict = errors.New("async idempotency identity has different content")
)

type EnqueueResult struct {
	ID      string
	Created bool
}

type Writer interface {
	EnqueueJob(context.Context, jobsv1.Enqueue) (EnqueueResult, error)
	EnqueueEvent(context.Context, eventsv1.Envelope) (EnqueueResult, error)
	CancelOwnedJob(context.Context, OwnedJobCancellation) (time.Time, error)
	ReleaseJobRetention(context.Context, jobsv1.RetentionRelease) (bool, error)
	CompleteDatabaseJob(context.Context, JobClaim) (time.Time, error)
	MarkExternalEffectStarted(
		context.Context, JobClaim, ExternalEffectRequestSpec,
	) (ExternalEffectPermit, error)
	FailJobBeforeEffect(
		context.Context, JobClaim, JobFailure,
	) (JobFailureResult, error)
	BeginAuthorizedHTTPIdempotency(
		context.Context, idempotency.Request,
	) (HTTPIdempotencyDecision, error)
	CompleteHTTPIdempotency(
		context.Context, HTTPIdempotencyPermit, idempotency.Response,
	) (idempotency.Receipt, error)
}

package catalog

import (
	"context"
	"errors"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/authority"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (processor *workerProcessor) persistResumeFailure(
	ctx context.Context,
	scope platformtx.TenantScope,
	claim asyncpostgres.JobClaim,
	snapshot lifecyclepostgres.JobSnapshot,
	cause error,
) error {
	evidence, err := authority.NewWorkerEvidence(
		snapshot.ActorID, scope.TenantID(), "catalog.operation.resume", "catalog.operation.resume",
	)
	if err != nil {
		return settleLost(asyncUnavailable())
	}
	authorized, err := authority.WithEvidence(ctx, evidence)
	if err != nil {
		return settleLost(asyncUnavailable())
	}
	operationID, err := uuid.Parse(snapshot.Operation.ID.String())
	if err != nil || operationID.Version() != uuid.Version(7) {
		return settleLost(asyncUnavailable())
	}
	if !lifecycleapp.ResumeFailureDurable(cause) &&
		claim.Attempt() < claim.MaximumAttempts() {
		state, failErr := processor.async.FailBeforeEffect(
			authorized, scope, claim, cause,
		)
		if failErr != nil {
			return settleLost(failErr)
		}
		if state != jobsv1.StateDead {
			return nil
		}
		cause = lifecycleapp.ErrResumeInterrupted
	}
	if !lifecycleapp.ResumeFailureDurable(cause) {
		cause = lifecycleapp.ErrResumeInterrupted
	}
	operation, err := processor.lifecycle.PersistResumeOutcome(
		authorized, scope, operationID, cause,
	)
	if err != nil {
		return settleLost(err)
	}
	return settleLost(processor.async.Settle(
		authorized, scope, claim, operation, uuid.Nil, time.Time{},
	))
}

func settleLost(err error) error {
	if err == nil || errors.Is(err, asyncpostgres.ErrJobLeaseLost) {
		return err
	}
	return nil
}

package catalog

import (
	"context"
	"errors"
	"time"

	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/authority"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (processor *workerProcessor) Process(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	snapshot, err := processor.claimSnapshot(ctx, claim)
	if err != nil {
		return err
	}
	tenantID, _ := claim.TenantID()
	request, err := processor.scopeRequest(ctx, tenantID, snapshot)
	if err != nil {
		return err
	}
	return runResumeScopes(ctx, processor.scopes, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			return processor.processInScope(scoped, scope, claim, snapshot)
		},
		func(scoped context.Context, scope platformtx.TenantScope, cause error) error {
			return processor.persistResumeFailure(
				scoped, scope, claim, snapshot, cause,
			)
		},
	)
}

func runResumeScopes(
	ctx context.Context,
	scopes WorkerScopeCoordinator,
	request platformtx.ScopeRequest,
	resume func(context.Context, platformtx.TenantScope) error,
	persist func(context.Context, platformtx.TenantScope, error) error,
) error {
	err := scopes.InTenant(ctx, request, resume)
	var failure *resumeFailure
	if !errors.As(err, &failure) {
		return err
	}
	return scopes.InTenant(ctx, request, func(
		scoped context.Context, scope platformtx.TenantScope,
	) error {
		return persist(scoped, scope, failure.cause)
	})
}

func (processor *workerProcessor) processInScope(
	ctx context.Context,
	scope platformtx.TenantScope,
	claim asyncpostgres.JobClaim,
	snapshot lifecyclepostgres.JobSnapshot,
) error {
	evidence, err := authority.NewWorkerEvidence(
		snapshot.ActorID, scope.TenantID(), "catalog.operation.resume", "catalog.operation.resume",
	)
	if err != nil {
		return asyncUnavailable()
	}
	authorized, err := authority.WithEvidence(ctx, evidence)
	if err != nil {
		return asyncUnavailable()
	}
	operation := snapshot.Operation
	operationID, err := uuid.Parse(operation.ID.String())
	if err != nil || operationID.Version() != uuid.Version(7) {
		return asyncUnavailable()
	}
	if !lifecycleapp.IsTerminalResume(operation) {
		result, resumeErr := processor.lifecycle.Resume(authorized, scope, operationID)
		if resumeErr != nil {
			return &resumeFailure{cause: resumeErr}
		} else {
			operation = result.Operation
			return processor.settleResult(authorized, scope, claim, result)
		}
	}
	return processor.async.Settle(
		authorized, scope, claim, operation, uuid.Nil, time.Time{},
	)
}

type resumeFailure struct{ cause error }

func (failure *resumeFailure) Error() string {
	if failure == nil || failure.cause == nil {
		return "catalog resume failed"
	}
	return failure.cause.Error()
}

func (failure *resumeFailure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.cause
}

func (processor *workerProcessor) settleResult(
	ctx context.Context,
	scope platformtx.TenantScope,
	claim asyncpostgres.JobClaim,
	result lifecycleapp.ResumeResult,
) error {
	if lifecycleapp.IsTerminalResume(result.Operation) {
		return processor.async.Settle(ctx, scope, claim, result.Operation, uuid.Nil, time.Time{})
	}
	nextID, err := processor.values.New()
	if err != nil {
		return asyncUnavailable()
	}
	runAt := result.NextAt
	now := processor.values.Now()
	if runAt.IsZero() || runAt.Before(now) {
		runAt = now
	}
	minimum := claim.StartedAt().UTC().Add(time.Microsecond)
	if runAt.Before(minimum) {
		runAt = minimum
	}
	return processor.async.Settle(ctx, scope, claim, result.Operation, nextID, runAt)
}

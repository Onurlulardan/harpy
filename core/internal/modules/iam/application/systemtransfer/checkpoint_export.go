package systemtransfer

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func (worker *ExportWorker) checkLease(ctx context.Context, claim ExportJobClaim) error {
	if err := worker.lease.CheckpointExportLease(ctx, claim); err != nil {
		if errors.Is(err, ErrExportCancelled) || errors.Is(err, ErrExportLeaseLost) {
			return err
		}
		return errors.Join(ErrExportUnavailable, err)
	}
	return nil
}

func (worker *ExportWorker) claimExport(
	ctx context.Context,
	attempt ExportAttempt,
) (result ExportClaimResult, err error) {
	err = worker.scopes.InFreshExportCheckpoint(ctx, attempt,
		func(scoped context.Context, scope ExportScope) error {
			store, authorizeErr := worker.authorizedExportStore(scoped, scope, attempt)
			if authorizeErr != nil {
				return authorizeErr
			}
			result, authorizeErr = store.AcquireOutbound(scoped, attempt)
			return authorizeErr
		})
	if err != nil {
		return ExportClaimResult{}, err
	}
	if (result.State != ExportRunning && result.State != ExportReady) ||
		(result.State == ExportReady && !result.Replay) {
		return ExportClaimResult{}, ErrExportStateConflict
	}
	return result, nil
}

func (worker *ExportWorker) checkpointExport(
	ctx context.Context,
	claim ExportJobClaim,
	attempt ExportAttempt,
	progress ExportProgress,
) error {
	if err := worker.checkLease(ctx, claim); err != nil {
		return err
	}
	return worker.scopes.InFreshExportCheckpoint(ctx, attempt,
		func(scoped context.Context, scope ExportScope) error {
			store, err := worker.authorizedExportStore(scoped, scope, attempt)
			if err != nil {
				return err
			}
			return store.CheckpointExport(scoped, attempt, progress)
		})
}

func (worker *ExportWorker) authorizedExportStore(
	ctx context.Context,
	scope ExportScope,
	attempt ExportAttempt,
) (ExportWorkerStore, error) {
	if scope.TenantID() != attempt.TenantID || scope.ActorID() != attempt.ActorID {
		return nil, ErrExportAuthorizationRevoked
	}
	evidence, err := worker.authorizer.AuthorizeExportCheckpoint(
		ctx, scope.IAMScope(), attempt,
	)
	if err != nil || exportEvidenceMatches(evidence, attempt) != nil {
		return nil, errors.Join(ErrExportAuthorizationRevoked, err)
	}
	store, err := worker.stores.BindExportWorker(ctx, scope.IAMScope())
	if err != nil || nilDependency(store) {
		return nil, errors.Join(ErrExportUnavailable, err)
	}
	return store, nil
}

func exportEvidenceMatches(
	evidence publicresource.ActorEvidence,
	attempt ExportAttempt,
) error {
	tenantID, tenantErr := evidence.TenantID()
	actorID, actorErr := evidence.ActorID()
	token, authorization, versionErr := evidence.Versions()
	if evidence.Validate() != nil || tenantErr != nil || actorErr != nil ||
		versionErr != nil || tenantID != publicresource.ID(attempt.TenantID) ||
		actorID != publicresource.ID(attempt.ActorID) || token != attempt.TokenVersion ||
		authorization != attempt.AuthorizationVersion {
		return ErrExportAuthorizationRevoked
	}
	return nil
}

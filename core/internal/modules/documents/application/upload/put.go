package upload

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Put(ctx context.Context, runner ScopeRunner,
	request PutRequest,
) (ReadyResult, error) {
	return application.put(ctx, runner, request, OperationPut, "")
}

func (application *Application) put(ctx context.Context, runner ScopeRunner,
	request PutRequest, operation, permission string,
) (ReadyResult, error) {
	if application == nil || foundation.NilPort(runner) || request.Content == nil ||
		!request.UploadID.UUIDv7() || !request.LeaseOwnerID.UUIDv7() ||
		!platformtx.NetworkAllowed(ctx) {
		return ReadyResult{}, ErrInvalid
	}
	if result, terminal, err := application.preflightPut(ctx, runner, request.UploadID,
		operation, permission); terminal || err != nil {
		return result, err
	}
	plan, err := application.stageAttempt(ctx, runner, request, operation, permission)
	if err != nil {
		return ReadyResult{}, err
	}
	receipt, failure := application.effects.Execute(ctx, plan, request.Content)
	if failure != nil {
		return ReadyResult{}, application.settleFailure(ctx, runner, *failure,
			operation, permission)
	}
	if !validEffectReceipt(plan, receipt) {
		return ReadyResult{}, application.settleFailure(ctx, runner, EffectFailure{
			Plan: plan, Kind: FailureInvalid, StableCode: "DOCUMENT_EFFECT_RECEIPT_INVALID",
			Object: &receipt.Object, OccurredAt: application.clock.Now()}, operation, permission)
	}
	if err = application.persistEffect(ctx, runner, receipt, operation, permission); err != nil {
		return ReadyResult{}, err
	}
	return application.settleReady(ctx, runner, receipt, operation, permission)
}

func (application *Application) preflightPut(ctx context.Context, runner ScopeRunner,
	uploadID domain.ID, operation, permission string,
) (ReadyResult, bool, error) {
	var snapshot Snapshot
	err := runner.InCommand(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		_, current, bindErr := application.bindPut(scoped, scope, uploadID,
			operation, permission)
		snapshot = current
		return bindErr
	})
	if err != nil {
		return ReadyResult{}, true, mapStoreError(err)
	}
	return replayResult(snapshot)
}

func (application *Application) stageAttempt(ctx context.Context, runner ScopeRunner,
	request PutRequest, operation, permission string,
) (AttemptPlan, error) {
	var plan AttemptPlan
	err := runner.InCommand(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := application.bindPut(scoped, scope, request.UploadID,
			operation, permission)
		if bindErr != nil {
			return bindErr
		}
		ids, idErr := application.nextIDs(2)
		if idErr != nil {
			return idErr
		}
		plan, bindErr = store.Begin(scoped, request.UploadID, request.LeaseOwnerID, ids[0])
		if bindErr != nil {
			return mapStoreError(bindErr)
		}
		if bindErr = application.billing.StartEffect(scoped, scope,
			plan.BillingReservationID, plan.EffectReceiptID); bindErr != nil {
			return mapBillingError(bindErr)
		}
		payload, _ := json.Marshal(struct {
			UploadID   string `json:"upload_id"`
			Generation uint64 `json:"generation"`
		}{UploadID: plan.UploadID.String(), Generation: plan.Generation})
		return application.jobs.EnqueueReconciliation(scoped, scope, ids[1],
			plan.UploadID, plan.Generation, application.clock.Now().UTC().Add(5*time.Minute),
			payload)
	})
	if err != nil {
		return AttemptPlan{}, mapStoreError(err)
	}
	return plan, nil
}

func (application *Application) persistEffect(ctx context.Context, runner ScopeRunner,
	receipt EffectReceipt, operation, permission string,
) error {
	return mapStoreError(runner.InCommand(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, err := application.bind(scoped, scope, Command,
			operation, putPermission(receipt.Plan, permission))
		if err != nil {
			return err
		}
		return store.PersistEffect(scoped, receipt)
	}))
}

func (application *Application) settleFailure(ctx context.Context, runner ScopeRunner,
	failure EffectFailure, operation, permission string,
) error {
	ids, err := application.nextIDs(2)
	if err != nil {
		return err
	}
	deletionID := domain.ID{}
	if failure.Object != nil {
		deletionID = ids[0]
	}
	err = runner.InCommand(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := application.bind(scoped, scope, Command,
			operation, putPermission(failure.Plan, permission))
		if bindErr != nil {
			return bindErr
		}
		if bindErr = store.PersistFailure(scoped, failure, deletionID); bindErr != nil {
			return bindErr
		}
		if failure.Object != nil {
			payload, _ := json.Marshal(struct {
				DeletionID string `json:"deletion_id"`
			}{DeletionID: deletionID.String()})
			return application.jobs.EnqueueDeletion(scoped, scope, deletionID,
				deletionID, payload)
		}
		if failure.Object == nil && failure.Kind != FailureTransient {
			return application.billing.Release(scoped, scope,
				failure.Plan.BillingReservationID, ids[1])
		}
		return nil
	})
	if err != nil {
		return mapStoreError(err)
	}
	switch failure.Kind {
	case FailureRejected, FailureInvalid:
		return ErrContentRejected
	default:
		return ErrScanUnavailable
	}
}

func validEffectReceipt(plan AttemptPlan, receipt EffectReceipt) bool {
	declaration := domain.FileDeclaration{Filename: plan.OriginalName,
		DeclaredMIME: plan.DeclaredMIME, DetectedKind: receipt.DetectedKind,
		Size: plan.ExpectedBytes}
	return receipt.Plan == plan && receipt.Object.Key == plan.ObjectKey &&
		receipt.Object.Generation == plan.Generation &&
		receipt.Object.PlaintextBytes == plan.ExpectedBytes &&
		receipt.Object.PlaintextSHA256 == plan.ExpectedContentSHA256 &&
		receipt.DetectedKind == plan.DeclaredKind && receipt.DetectedMIME == plan.DeclaredMIME &&
		domain.ValidateFile(declaration) == nil && receipt.ScanRevision != "" &&
		foundation.ValidInstant(receipt.ScannedAt)
}

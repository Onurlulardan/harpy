package upload

import (
	"context"
	"encoding/json"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) settleReady(ctx context.Context, runner ScopeRunner,
	receipt EffectReceipt, operation, permission string,
) (ReadyResult, error) {
	ids, err := application.nextIDs(3)
	if err != nil {
		return ReadyResult{}, err
	}
	var result ReadyResult
	var outcome error
	err = runner.InCommand(ctx, operation, func(scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		store, _, bindErr := application.bind(scoped, scope, Command,
			operation, putPermission(receipt.Plan, permission))
		if bindErr != nil {
			return bindErr
		}
		decision, prepareErr := store.PrepareFinalize(scoped, receipt)
		if prepareErr != nil {
			return prepareErr
		}
		if decision.Conflict || decision.Cancelled {
			code := "DOCUMENT_VERSION_CONFLICT"
			if decision.Cancelled {
				code = "DOCUMENT_UPLOAD_CANCELLED"
			}
			failure := EffectFailure{Plan: receipt.Plan, Kind: FailureInvalid,
				StableCode: code, Object: &receipt.Object, OccurredAt: application.clock.Now()}
			if failureErr := store.PersistFailure(scoped, failure, ids[0]); failureErr != nil {
				return failureErr
			}
			payload, _ := json.Marshal(struct {
				DeletionID string `json:"deletion_id"`
			}{DeletionID: ids[0].String()})
			if enqueueErr := application.jobs.EnqueueDeletion(scoped, scope, ids[0],
				ids[0], payload); enqueueErr != nil {
				return enqueueErr
			}
			outcome = ErrConflict
			return nil
		}
		if !decision.Ready {
			return ErrUnavailable
		}
		if bindErr = application.billing.Commit(scoped, scope,
			receipt.Plan.BillingReservationID, ids[0],
			receipt.Object.CiphertextBytes); bindErr != nil {
			return bindErr
		}
		result, bindErr = store.PersistFinalization(scoped, receipt, ids[1], ids[2])
		return bindErr
	})
	if err != nil {
		return ReadyResult{}, mapStoreError(err)
	}
	if outcome != nil {
		return ReadyResult{}, outcome
	}
	if result.Status != 201 || result.DocumentID != receipt.Plan.DocumentID ||
		result.VersionID != receipt.Plan.CandidateVersionID {
		return ReadyResult{}, ErrUnavailable
	}
	return result, nil
}

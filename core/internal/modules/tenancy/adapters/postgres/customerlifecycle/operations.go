package customerlifecycle

import (
	"context"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

func (store *Store) Accept(ctx context.Context, request lifecycle.AsyncStoreRequest) (
	lifecycle.AsyncStoreResult, error,
) {
	if !store.usable(ctx) {
		return lifecycle.AsyncStoreResult{}, lifecycle.ErrInvalid
	}
	var outcome string
	operation, err := scanAcceptedOperation(store.queryRow(ctx, acceptSQL,
		request.EnvironmentID, request.ControlTenantID, request.ActorID, request.TargetTenantID,
		request.Kind, request.ExpectedRevision, request.IdempotencyKey, request.RequestSHA256[:],
		nullProofUUID(request.Proof.RetentionReceiptID), nullProofUUID(request.Proof.BackupReceiptID),
		nullProofUUID(request.Proof.ContributorReceiptID), nullDigest(request.Proof.EvidenceSHA256),
		nullTime(request.Proof.VerifiedAt), request.Now), store.classify, &outcome)
	if err != nil {
		return lifecycle.AsyncStoreResult{}, err
	}
	return lifecycle.AsyncStoreResult{Operation: operation, Outcome: outcome}, nil
}

func scanAcceptedOperation(row Row, classify func(error) error, prefix ...any) (
	lifecycle.Operation, error,
) {
	var id, controlID, targetID *uuid.UUID
	var kind, state, failureCode *string
	var revision *int64
	var replay *bool
	var createdAt, updatedAt, completedAt *time.Time
	targets := append(prefix, &id, &controlID, &targetID, &kind, &state, &revision,
		&replay, &createdAt, &updatedAt, &completedAt, &failureCode)
	if err := row.Scan(targets...); err != nil {
		return lifecycle.Operation{}, classify(err)
	}
	if id == nil {
		return lifecycle.Operation{}, nil
	}
	if controlID == nil || targetID == nil || kind == nil || state == nil ||
		revision == nil || replay == nil || createdAt == nil || updatedAt == nil {
		return lifecycle.Operation{}, lifecycle.ErrUnavailable
	}
	result := lifecycle.Operation{ID: *id, ControlTenantID: *controlID,
		TargetTenantID: *targetID, Kind: *kind, State: *state,
		Revision: *revision, Replay: *replay,
		CreatedAt: createdAt.UTC(), UpdatedAt: updatedAt.UTC()}
	if completedAt != nil {
		result.CompletedAt = completedAt.UTC()
	}
	if failureCode != nil {
		result.FailureCode = *failureCode
	}
	return result, nil
}

func (store *Store) ReadOperation(ctx context.Context, controlID, operationID uuid.UUID) (
	lifecycle.Operation, error,
) {
	if !store.usable(ctx) || controlID == uuid.Nil || operationID == uuid.Nil {
		return lifecycle.Operation{}, lifecycle.ErrInvalid
	}
	result, _, err := scanOperation(
		store.queryRow(ctx, readOperationSQL, controlID, operationID), store.classify,
	)
	return result, err
}

func nullProofUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func nullDigest(value [32]byte) any {
	if value == [32]byte{} {
		return nil
	}
	return value[:]
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

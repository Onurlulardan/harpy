package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func (store *WorkerStore) LoadReceipts(ctx context.Context,
	scope tenantScope, evidence publicresource.ActorEvidence,
	attempt systemtransferapp.Attempt, indexes []int64,
) (map[int64]systemtransferapp.RowReceipt, error) {
	if attempt.Validate() != nil || len(indexes) < 1 || len(indexes) > 1000 {
		return nil, systemtransferapp.ErrInvalidAttempt
	}
	token, authorization, err := workerEvidence(scope, evidence, attempt)
	if err != nil {
		return nil, err
	}
	bound, err := store.bind(ctx, scope)
	if err != nil {
		return nil, err
	}
	schema, err := bound.database.CurrentSchema(ctx)
	if err != nil {
		return nil, systemtransferapp.ErrWorkerUnavailable
	}
	rows, err := bound.database.Query(ctx, readImportWorkerReceiptsSQL,
		schema, attempt.TenantID, attempt.ActorID, attempt.ID, attempt.JobID,
		attempt.AttemptID, attempt.LeaseOwnerID, attempt.LeaseGeneration,
		string(attempt.Resource), token, authorization, indexes)
	if err != nil {
		return nil, mapWorkerError(err)
	}
	defer rows.Close()
	result := make(map[int64]systemtransferapp.RowReceipt, len(indexes))
	for rows.Next() {
		receipt, scanErr := scanWorkerReceipt(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if _, duplicate := result[receipt.RowIndex]; duplicate {
			return nil, systemtransferapp.ErrWorkerState
		}
		result[receipt.RowIndex] = receipt
	}
	if err = rows.Err(); err != nil {
		return nil, mapWorkerError(err)
	}
	return result, nil
}

func scanWorkerReceipt(row interface{ Scan(...any) error },
) (systemtransferapp.RowReceipt, error) {
	var index int64
	var revision *int64
	var outcome string
	var ownerKind, errorCode *string
	var ownerID *uuid.UUID
	var source, result []byte
	err := row.Scan(&index, &outcome, &ownerKind, &ownerID, &revision,
		&errorCode, &source, &result)
	if err != nil {
		return systemtransferapp.RowReceipt{}, mapWorkerError(err)
	}
	owner, code, id, ownerRevision := "", "", uuid.Nil, int64(0)
	if ownerKind != nil {
		owner = *ownerKind
	}
	if errorCode != nil {
		code = *errorCode
	}
	if ownerID != nil {
		id = *ownerID
	}
	if revision != nil {
		ownerRevision = *revision
	}
	sourceDigest, sourceOK := workerDigest(source)
	resultDigest, resultOK := workerDigest(result)
	receipt, err := systemtransferapp.NewRowReceipt(
		index, outcome, owner, id, ownerRevision, code, sourceDigest, resultDigest)
	if err != nil || !sourceOK || !resultOK {
		return systemtransferapp.RowReceipt{}, systemtransferapp.ErrWorkerState
	}
	return receipt, nil
}

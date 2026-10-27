package systemtransfer

import (
	"context"
	"encoding/hex"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type receiptWire struct {
	RowIndex      int64      `json:"row_index"`
	Outcome       string     `json:"outcome"`
	OwnerKind     *string    `json:"owner_kind"`
	OwnerID       *uuid.UUID `json:"owner_id"`
	OwnerRevision *int64     `json:"owner_revision"`
	ErrorCode     *string    `json:"error_code"`
	SourceSHA256  string     `json:"source_sha256"`
	ResultSHA256  string     `json:"result_sha256"`
}

func (store *WorkerStore) CommitReceipts(ctx context.Context,
	scope tenantScope, evidence publicresource.ActorEvidence,
	attempt systemtransferapp.Attempt, receipts []systemtransferapp.RowReceipt,
	sourceBytes int64,
) error {
	if attempt.Validate() != nil || len(receipts) < 1 || len(receipts) > 1000 ||
		sourceBytes < 2 {
		return systemtransferapp.ErrInvalidAttempt
	}
	token, authorization, err := workerEvidence(scope, evidence, attempt)
	if err != nil {
		return err
	}
	document, err := encodeReceiptWires(receipts)
	if err != nil {
		return err
	}
	bound, err := store.bind(ctx, scope)
	if err != nil {
		return err
	}
	schema, err := bound.database.CurrentSchema(ctx)
	if err != nil {
		return systemtransferapp.ErrWorkerUnavailable
	}
	var progress [7]int64
	err = bound.database.QueryRow(ctx, commitImportWorkerBatchSQL,
		schema, attempt.TenantID, attempt.ActorID, attempt.ID, attempt.JobID,
		attempt.AttemptID, attempt.LeaseOwnerID, attempt.LeaseGeneration,
		string(attempt.Resource), token, authorization, document, sourceBytes,
	).Scan(&progress[0], &progress[1], &progress[2], &progress[3],
		&progress[4], &progress[5], &progress[6])
	return mapWorkerError(err)
}

func encodeReceiptWires(receipts []systemtransferapp.RowReceipt) ([]byte, error) {
	wires := make([]receiptWire, len(receipts))
	for index, receipt := range receipts {
		if !receipt.Valid() {
			return nil, systemtransferapp.ErrInvalidReceipt
		}
		wire := receiptWire{RowIndex: receipt.RowIndex, Outcome: receipt.Outcome,
			SourceSHA256: hex.EncodeToString(receipt.SourceSHA256[:]),
			ResultSHA256: hex.EncodeToString(receipt.ResultSHA256[:])}
		if receipt.OwnerKind != "" {
			wire.OwnerKind = &receipt.OwnerKind
		}
		if receipt.OwnerID != uuid.Nil {
			wire.OwnerID = &receipt.OwnerID
		}
		if receipt.OwnerRevision > 0 {
			wire.OwnerRevision = &receipt.OwnerRevision
		}
		if receipt.ErrorCode != "" {
			wire.ErrorCode = &receipt.ErrorCode
		}
		wires[index] = wire
	}
	document, err := json.Marshal(wires)
	if err != nil || len(document) > 1<<20 {
		return nil, systemtransferapp.ErrInvalidReceipt
	}
	return document, nil
}

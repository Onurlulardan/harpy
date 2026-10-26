package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) LoadOperation(
	ctx context.Context,
	operationID string,
) (ports.OperationRecord, error) {
	var row operationRow
	err := store.withBound(ctx, transactionSnapshotQuery, func(bound boundStore) error {
		return store.mapDatabaseError(bound.database.QueryRow(ctx, getOperationSQL, operationID).Scan(
			&row.operationID, &row.purpose, &row.bindingID, &row.scopeKind, &row.scopeID,
			&row.generation, &row.previousHead, &row.stateSHA, &row.stateBytes,
			&row.proposalSHA, &row.proposalBytes, &row.compensatesAck, &row.status,
			&row.ackSHA, &row.ackBytes, &row.registrySequence, &row.completedAt,
		))
	})
	if err != nil {
		return ports.OperationRecord{}, fmt.Errorf("get platform authority operation: %w", err)
	}
	return row.operation()
}

type operationRow struct {
	operationID, purpose, bindingID, scopeKind, scopeID string
	generation                                          int64
	previousHead, compensatesAck, ackSHA                *string
	stateSHA, proposalSHA, status                       string
	stateBytes, proposalBytes, ackBytes                 []byte
	registrySequence                                    *int64
	completedAt                                         *time.Time
}

func (row operationRow) operation() (ports.OperationRecord, error) {
	state, err := artifact.ParseState(row.stateBytes)
	if err != nil || state.SHA256 != row.stateSHA {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	proposal, err := artifact.ParseSignedProposal(row.proposalBytes)
	if err != nil || proposal.SHA256 != row.proposalSHA ||
		proposal.Value.OperationID != row.operationID || proposal.Value.Generation != row.generation {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	result := ports.OperationRecord{
		OperationID: row.operationID, Purpose: row.purpose, State: state,
		Proposal: proposal, Status: ports.OperationStatus(row.status),
	}
	if len(row.ackBytes) > 0 {
		result.Ack, err = artifact.ParseAck(row.ackBytes)
		if err != nil || row.ackSHA == nil || result.Ack.SHA256 != *row.ackSHA {
			return ports.OperationRecord{}, ports.ErrConflict
		}
	}
	if row.completedAt != nil {
		result.CompletedAt = row.completedAt.UTC()
	}
	return result, nil
}

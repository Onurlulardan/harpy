package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) PrepareOperation(
	ctx context.Context,
	record ports.OperationRecord,
) (ports.OperationRecord, error) {
	if !validOperationRecord(record) || record.Status != ports.StatusProposalSigned {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	issuedAt, _ := parseOperationInstant(record.Proposal.Value.IssuedAt)
	expiresAt, _ := parseOperationInstant(record.Proposal.Value.ExpiresAt)
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	value := record.Proposal.Value
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, prepareSQL,
			record.OperationID, value.RegistryBindingID, record.Purpose,
			string(value.ScopeKind), value.ScopeID, value.Generation,
			optionalText(value.PreviousHeadSHA256), record.State.SHA256,
			record.State.Bytes, record.Proposal.SHA256, record.Proposal.Bytes,
			optionalText(value.CompensatesAckSHA256), issuedAt, expiresAt,
			value.SignerKID, now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.OperationRecord{}, fmt.Errorf("prepare platform authority operation: %w", err)
	}
	return store.LoadOperation(ctx, record.OperationID)
}

package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) PersistAcknowledgement(
	ctx context.Context,
	operationID string,
	ack artifact.VerifiedAck,
) (ports.OperationRecord, error) {
	operation, err := store.LoadOperation(ctx, operationID)
	if err != nil || !ackMatchesOperation(ack, operation) {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	value := ack.Value
	err = store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, recordAckSQL,
			operationID, value.RegistryBindingID, string(value.ScopeKind), value.ScopeID,
			value.Generation, optionalText(value.PreviousHeadSHA256), value.StateSHA256,
			value.ProposalSHA256, ack.SHA256, ack.Bytes, value.RegistrySequence, now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.OperationRecord{}, fmt.Errorf("record platform authority ack: %w", err)
	}
	return store.LoadOperation(ctx, operationID)
}

func ackMatchesOperation(ack artifact.VerifiedAck, operation ports.OperationRecord) bool {
	value := ack.Value
	proposal := operation.Proposal.Value
	return value.OperationID == operation.OperationID &&
		value.RegistryBindingID == proposal.RegistryBindingID &&
		value.ScopeKind == proposal.ScopeKind && value.ScopeID == proposal.ScopeID &&
		value.Generation == proposal.Generation &&
		value.ProposalSHA256 == operation.Proposal.SHA256 &&
		value.StateSHA256 == operation.State.SHA256 &&
		optionalDigestEqual(value.PreviousHeadSHA256, proposal.PreviousHeadSHA256)
}

func optionalDigestEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

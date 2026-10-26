package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) CommitAcknowledged(
	ctx context.Context,
	operationID string,
) (ports.CommitResult, error) {
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.CommitResult{}, ports.ErrConflict
	}
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, commitSQL, operationID, now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.CommitResult{}, fmt.Errorf("commit platform authority ack: %w", err)
	}
	operation, err := store.LoadOperation(ctx, operationID)
	if err != nil {
		return ports.CommitResult{}, err
	}
	head, err := store.LoadHead(ctx, ports.HeadKey{
		RegistryBindingID: operation.Proposal.Value.RegistryBindingID,
		ScopeKind:         operation.Proposal.Value.ScopeKind,
		ScopeID:           operation.Proposal.Value.ScopeID,
	})
	if err != nil {
		return ports.CommitResult{}, err
	}
	return ports.CommitResult{Operation: operation, Head: head}, nil
}

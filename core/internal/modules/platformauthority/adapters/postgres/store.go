package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type Store struct {
	transactions transactionRunner
	clock        ports.Clock
	rowNotFound  func(error) bool
}

type boundStore struct {
	database boundDatabase
}

func (store *Store) deliveryDisposition(
	ctx context.Context,
	statement string,
	claim ports.DeliveryClaim,
	extra ...any,
) error {
	if claim.OperationID == "" || claim.LeaseOwnerID == "" ||
		claim.LeaseGeneration < 1 {
		return ports.ErrConflict
	}
	arguments := []any{
		claim.OperationID, claim.LeaseOwnerID, claim.LeaseGeneration,
	}
	arguments = append(arguments, extra...)
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, statement, arguments...)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("record platform authority delivery disposition: %w", err)
	}
	return nil
}

func (store *Store) withBound(
	ctx context.Context,
	mode transactionMode,
	callback func(boundStore) error,
) error {
	if store == nil || store.transactions == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("platform authority postgres store is unavailable")
	}
	return store.transactions(ctx, mode, func(database boundDatabase) error {
		return callback(boundStore{database: database})
	})
}

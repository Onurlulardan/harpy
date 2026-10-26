package restoreexecution

import (
	"context"
	"fmt"
)

type Store struct {
	transactions transactionRunner
	rowNotFound  func(error) bool
}

func (store *Store) withBound(
	ctx context.Context,
	callback func(boundDatabase) error,
) error {
	if store == nil || store.transactions == nil ||
		ctx == nil || ctx.Err() != nil || callback == nil {
		return fmt.Errorf("restore execution store is unavailable")
	}
	return store.transactions(ctx, transactionCommand, callback)
}

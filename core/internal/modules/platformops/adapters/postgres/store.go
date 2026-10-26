package postgres

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
	mode transactionMode,
	callback func(boundDatabase) error,
) error {
	if store == nil || store.transactions == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("platform operations postgres store is unavailable")
	}
	return store.transactions(ctx, mode, callback)
}

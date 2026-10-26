package adoption

import (
	"context"
	"fmt"
)

type Row interface{ Scan(...any) error }
type Database interface {
	QueryRow(context.Context, string, ...any) Row
}
type TransactionRunner func(context.Context, func(Database) error) error

type Store struct{ run TransactionRunner }

func NewStoreForAdopter(run TransactionRunner) (*Store, error) {
	if run == nil {
		return nil, fmt.Errorf("usage adoption postgres profile is unavailable")
	}
	return &Store{run: run}, nil
}

func (store *Store) within(ctx context.Context, action func(Database) error) error {
	if store == nil || store.run == nil || ctx == nil || ctx.Err() != nil || action == nil {
		return fmt.Errorf("usage adoption postgres store is unavailable")
	}
	return store.run(ctx, action)
}

package postgres

import (
	"context"
	"fmt"
)

type TransactionMode string

const (
	CommandTransaction  TransactionMode = "command"
	SnapshotTransaction TransactionMode = "snapshot"
)

type Row interface{ Scan(...any) error }

type Database interface {
	QueryRow(context.Context, string, ...any) Row
}

type TransactionRunner func(context.Context, TransactionMode, func(Database) error) error

type Store struct {
	run         TransactionRunner
	rowNotFound func(error) bool
}

func NewStoreForProfile(run TransactionRunner, rowNotFound func(error) bool) (*Store, error) {
	if run == nil || rowNotFound == nil {
		return nil, fmt.Errorf("usage postgres profile is unavailable")
	}
	return &Store{run: run, rowNotFound: rowNotFound}, nil
}

func (store *Store) within(ctx context.Context, mode TransactionMode,
	execute func(Database) error,
) error {
	if store == nil || store.run == nil || ctx == nil || ctx.Err() != nil || execute == nil {
		return fmt.Errorf("usage postgres store is unavailable")
	}
	return store.run(ctx, mode, execute)
}

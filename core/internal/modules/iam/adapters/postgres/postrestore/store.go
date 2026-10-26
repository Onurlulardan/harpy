package postrestore

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/google/uuid"
)

type row interface {
	Scan(...any) error
}

type database interface {
	QueryRow(context.Context, string, ...any) row
}

type transactionMode uint8

const (
	snapshotQuery transactionMode = iota + 1
	transactionCommand
)

type Store struct {
	withinGlobal func(context.Context, transactionMode, func(database) error) error
	withinTenant func(context.Context, transactionMode, uuid.UUID, func(database) error) error
	rowNotFound  func(error) bool
}

func (store *Store) global(
	ctx context.Context,
	mode transactionMode,
	callback func(database) error,
) error {
	return store.withinGlobal(ctx, mode, callback)
}

func (store *Store) tenant(
	ctx context.Context,
	mode transactionMode,
	tenantID uuid.UUID,
	callback func(database) error,
) error {
	return store.withinTenant(ctx, mode, tenantID, callback)
}

var _ postrestore.Store = (*Store)(nil)

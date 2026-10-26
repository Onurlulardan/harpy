package postgres

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type RetentionStore struct{ database Database }

func NewRetentionStore(database Database) (*RetentionStore, error) {
	if nilRetentionDatabase(database) {
		return nil, fmt.Errorf("usage retention database is required")
	}
	return &RetentionStore{database: database}, nil
}

func (store *RetentionStore) Prune(
	ctx context.Context,
	request application.RetentionPrune,
) (int, error) {
	if store == nil || nilRetentionDatabase(store.database) || ctx == nil ||
		ctx.Err() != nil || request.Validate() != nil {
		return 0, domain.ErrForbidden
	}
	var deleted int
	err := store.database.QueryRow(ctx, retentionPruneSQL, request.ReceiptID,
		request.EnvironmentID, request.DeleteBefore, request.BatchLimit,
		request.ObservedAt).Scan(&deleted)
	if err != nil {
		return 0, normalizePostgresError(err, nil)
	}
	if deleted < 0 || deleted > request.BatchLimit {
		return 0, domain.ErrUnavailable
	}
	return deleted, nil
}

func nilRetentionDatabase(database Database) bool {
	if database == nil {
		return true
	}
	value := reflect.ValueOf(database)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

var _ application.RetentionStore = (*RetentionStore)(nil)

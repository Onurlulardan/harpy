package postgres

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	"github.com/jackc/pgx/v5"
)

var (
	ErrRetentionStoreUnavailable = errors.New("async retention store is unavailable")
	ErrRetentionResultInvalid    = errors.New("async retention result is invalid")
)

type retentionDatabase interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type retentionRepository struct{ database retentionDatabase }

var _ retention.Store = (*retentionRepository)(nil)

func (repository *retentionRepository) valid() bool {
	return repository != nil && repository.database != nil
}

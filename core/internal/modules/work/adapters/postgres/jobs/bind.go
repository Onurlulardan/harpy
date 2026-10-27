package jobs

import (
	"context"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func NewBoundStore(database platformtx.JobBoundExecutor) (jobapp.Store, error) {
	store := &Store{database: databaseAccess{
		execute: func(ctx context.Context, statement string, arguments ...any) (int64, error) {
			result, err := database.Execute(ctx, statement, arguments...)
			return result.RowsAffected(), err
		},
		query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
			return database.Query(ctx, statement, arguments...)
		},
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return database.QueryRow(ctx, statement, arguments...)
		},
	}}
	return store, nil
}

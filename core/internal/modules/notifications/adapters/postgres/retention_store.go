package postgres

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
)

type RetentionRow interface{ Scan(...any) error }

type RetentionDatabaseAccess struct {
	QueryRow    func(context.Context, string, ...any) RetentionRow
	RowNotFound func(error) bool
}

type RetentionStore struct{ database RetentionDatabaseAccess }

func NewRetentionStore(database RetentionDatabaseAccess) (*RetentionStore, error) {
	if database.QueryRow == nil || database.RowNotFound == nil {
		return nil, fmt.Errorf("notification retention database is required")
	}
	return &RetentionStore{database: database}, nil
}

var _ retentionapp.Store = (*RetentionStore)(nil)

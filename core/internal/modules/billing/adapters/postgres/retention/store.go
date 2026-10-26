package retention

import (
	"context"
	"fmt"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/retention"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type store struct{ database DatabaseAccess }

func NewStore(database DatabaseAccess) (retentionapp.Store, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing retention database is required")
	}
	return &store{database: database}, nil
}

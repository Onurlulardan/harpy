package v1adoption

import (
	"context"
	"fmt"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type store struct{ database DatabaseAccess }

func NewStore(database DatabaseAccess) (billingv1adoptionapp.ScopeStore, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing V1 adoption database is required")
	}
	return &store{database: database}, nil
}

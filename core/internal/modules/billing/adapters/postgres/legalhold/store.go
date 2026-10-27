package legalhold

import (
	"context"
	"fmt"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	"github.com/google/uuid"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
	Allowed  func(context.Context) bool
}

type store struct {
	tenantID uuid.UUID
	database DatabaseAccess
}

func NewStore(
	tenantID uuid.UUID,
	database DatabaseAccess,
) (legalholdapp.Store, error) {
	if tenantID == uuid.Nil || database.QueryRow == nil || database.Allowed == nil {
		return nil, fmt.Errorf("billing legal hold database is required")
	}
	return &store{tenantID: tenantID, database: database}, nil
}

func (storage *store) Place(
	ctx context.Context,
	mutation legalholdapp.PlaceMutation,
) (legalholdapp.Snapshot, error) {
	return storage.Create(ctx, mutation)
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.tenantID != uuid.Nil &&
		storage.database.QueryRow != nil && storage.database.Allowed != nil &&
		ctx != nil && ctx.Err() == nil && storage.database.Allowed(ctx)
}

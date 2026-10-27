package licenseauthority

import (
	"context"
	"fmt"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

type Row interface{ Scan(...any) error }

type DatabaseAccess struct {
	QueryRow func(context.Context, string, ...any) Row
}

type Store interface {
	licenseapp.Store
	licenseapp.RestoredLicenseStore
	licenseapp.RuntimeStore
	licenseapp.RefreshTargetStore
	licenseapp.RestoreBindingStore
	licenseapp.RestoredLicenseReceiptStore
}

type store struct{ database DatabaseAccess }

var _ Store = (*store)(nil)

func NewStore(database DatabaseAccess) (Store, error) {
	if database.QueryRow == nil {
		return nil, fmt.Errorf("billing license authority database is required")
	}
	return &store{database: database}, nil
}

func (storage *store) usable(ctx context.Context) bool {
	return storage != nil && storage.database.QueryRow != nil &&
		ctx != nil && ctx.Err() == nil
}

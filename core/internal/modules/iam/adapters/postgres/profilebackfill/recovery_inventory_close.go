package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func (cursor *recoveryInventoryCursor) Close(ctx context.Context) error {
	if cursor == nil {
		return profilebackfillapp.ErrInvalid
	}
	if cursor.name == "" {
		return nil
	}
	if !cursor.usable(scopeUsable(ctx, cursor.scope)) {
		return profilebackfillapp.ErrInvalid
	}
	cursor.name = ""
	if err := cursor.store.closeRecoveryInventory(ctx); err != nil {
		return recoveryStoreError(err)
	}
	return nil
}

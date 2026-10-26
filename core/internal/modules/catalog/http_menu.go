package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) GetMenu(ctx context.Context, access cataloghttp.Access) (cataloghttp.Menu, error) {
	var result cataloghttp.Menu
	err := runtime.menuSnapshot(ctx, access,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.menu.Get(scoped, scope)
			if err == nil {
				result = cataloghttp.RepresentMenu(value)
			}
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) UpdateMenuOverride(ctx context.Context, access cataloghttp.Access,
	input cataloghttp.MenuOverrideInput,
) (cataloghttp.MenuOverrideReceipt, error) {
	var result cataloghttp.MenuOverrideReceipt
	err := runtime.catalogCommand(ctx, access, cataloghttp.PermissionMenuManage, "catalog.menu.update",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			receipt, err := runtime.menu.UpdateOverride(scoped, scope, menuapp.UpdateRequest{
				ItemID: input.ItemID, Label: input.Label, Order: input.Order, Visible: input.Visible,
				ExpectedRevision: input.ExpectedRevision, IdempotencyKey: input.IdempotencyKey})
			if err == nil {
				result = cataloghttp.MenuOverrideReceipt{ItemID: receipt.Override.ItemID.String(),
					Revision: receipt.Revision, Replayed: receipt.Replayed}
			}
			return err
		})
	return result, err
}

package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ReplaceMenu(ctx context.Context, access cataloghttp.Access,
	input cataloghttp.MenuReplaceInput,
) (cataloghttp.Menu, error) {
	var result cataloghttp.Menu
	err := runtime.catalogCommand(ctx, access, cataloghttp.PermissionMenuManage, "catalog.menu.update",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.menu.Replace(scoped, scope, menuapp.ReplaceRequest{
				ExpectedRevision: input.ExpectedRevision,
				Items:            replaceItems(input.Items),
				IdempotencyKey:   input.IdempotencyKey,
			})
			if err == nil {
				result = cataloghttp.RepresentMenu(value)
			}
			return err
		})
	return result, err
}

func replaceItems(items []cataloghttp.MenuReplaceItem) []menuapp.ReplaceItem {
	result := make([]menuapp.ReplaceItem, 0, len(items))
	for _, item := range items {
		result = append(result, menuapp.ReplaceItem{
			ItemID: item.ItemID, Kind: item.Kind, Label: item.Label,
			ParentID: item.ParentID, TargetID: item.TargetID,
			Order: item.Order, Visible: item.Visible,
		})
	}
	return result
}

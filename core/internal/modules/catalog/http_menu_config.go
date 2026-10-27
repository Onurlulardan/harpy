package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetActiveMenuConfig(
	ctx context.Context, access cataloghttp.Access,
) (*cataloghttp.MenuConfig, error) {
	var result *cataloghttp.MenuConfig
	err := runtime.menuSnapshot(ctx, access,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.menu.GetActiveConfig(scoped, scope)
			if err != nil {
				return err
			}
			if value != nil {
				represented := cataloghttp.RepresentMenuConfig(*value)
				result = &represented
			}
			return nil
		})
	return result, err
}

func (runtime *HTTPRuntime) CreateMenuConfig(ctx context.Context, access cataloghttp.Access,
	input cataloghttp.MenuConfigWriteInput,
) (cataloghttp.MenuConfig, error) {
	var result cataloghttp.MenuConfig
	err := runtime.catalogCommand(ctx, access, cataloghttp.PermissionMenuManage, "catalog.menu.update",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.menu.CreateConfig(scoped, scope, menuapp.ConfigWriteRequest{
				Name: input.Name, Config: input.Config, Active: true,
				IdempotencyKey: input.IdempotencyKey,
			})
			if err == nil {
				result = cataloghttp.RepresentMenuConfig(value)
			}
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) UpdateMenuConfig(ctx context.Context, access cataloghttp.Access,
	id uuid.UUID, input cataloghttp.MenuConfigWriteInput,
) (cataloghttp.MenuConfig, error) {
	var result cataloghttp.MenuConfig
	err := runtime.catalogCommand(ctx, access, cataloghttp.PermissionMenuManage, "catalog.menu.update",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.menu.UpdateConfig(scoped, scope, menuapp.ConfigWriteRequest{
				ID: id, Name: input.Name, Config: input.Config, Active: input.Active,
				IdempotencyKey: input.IdempotencyKey,
			})
			if err == nil {
				result = cataloghttp.RepresentMenuConfig(value)
			}
			return err
		})
	return result, err
}

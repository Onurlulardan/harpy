package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) ListInstallations(ctx context.Context, access cataloghttp.Access,
	limit int, after *uuid.UUID,
) ([]cataloghttp.Installation, *uuid.UUID, error) {
	var result []cataloghttp.Installation
	var next *uuid.UUID
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.installation.list",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			request := lifecycleapp.PageRequest{Limit: limit}
			if after != nil {
				request.After = *after
			}
			page, err := runtime.lifecycle.ListInstallations(scoped, scope, request)
			if err != nil {
				return err
			}
			result = make([]cataloghttp.Installation, 0, len(page.Items))
			for _, item := range page.Items {
				result = append(result, cataloghttp.RepresentInstallation(item))
			}
			if page.Next != uuid.Nil {
				value := page.Next
				next = &value
			}
			return nil
		})
	return result, next, err
}

func (runtime *HTTPRuntime) GetInstallation(ctx context.Context, access cataloghttp.Access,
	installationID uuid.UUID,
) (cataloghttp.Installation, error) {
	var result cataloghttp.Installation
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.installation.get",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.lifecycle.InstallationByID(scoped, scope, installationID)
			if err == nil {
				result = cataloghttp.RepresentInstallation(value)
			}
			return err
		})
	return result, err
}

func (runtime *HTTPRuntime) GetOperation(ctx context.Context, access cataloghttp.Access,
	operationID uuid.UUID,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead, "catalog.operation.get",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, err := runtime.lifecycle.ReadOperation(scoped, scope, operationID)
			if err == nil {
				result = cataloghttp.RepresentOperation(value, false)
			}
			return err
		})
	return result, err
}

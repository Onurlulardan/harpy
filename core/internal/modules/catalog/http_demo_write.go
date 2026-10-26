package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) InstallDemo(
	ctx context.Context,
	access cataloghttp.Access,
	installationID uuid.UUID,
	key string,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionDemoManage,
		"catalog.installation.demo.install",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			handle, startErr := runtime.lifecycle.InstallDemo(
				scoped, scope, lifecycleapp.DemoInstallRequest{
					InstallationID: installationID, IdempotencyKey: key,
				},
			)
			if startErr == nil {
				result = cataloghttp.RepresentHandle(handle)
			}
			return startErr
		})
	return result, err
}

func (runtime *HTTPRuntime) RemoveDemo(
	ctx context.Context,
	access cataloghttp.Access,
	installationID uuid.UUID,
	expectedPlanSHA256, key string,
) (cataloghttp.Operation, error) {
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionDemoManage,
		"catalog.installation.demo.remove",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			handle, startErr := runtime.lifecycle.RemoveDemo(
				scoped, scope, lifecycleapp.DemoRemoveRequest{
					InstallationID:     installationID,
					ExpectedPlanSHA256: expectedPlanSHA256,
					IdempotencyKey:     key,
				},
			)
			if startErr == nil {
				result = cataloghttp.RepresentHandle(handle)
			}
			return startErr
		})
	return result, err
}

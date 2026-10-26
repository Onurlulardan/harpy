package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) GetDemoStatus(
	ctx context.Context,
	access cataloghttp.Access,
	installationID uuid.UUID,
) (cataloghttp.DemoStatus, error) {
	var result cataloghttp.DemoStatus
	err := runtime.snapshot(ctx, access, cataloghttp.PermissionRead,
		"catalog.installation.demo.get",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, readErr := runtime.lifecycle.GetDemoStatus(
				scoped, scope, installationID,
			)
			if readErr == nil {
				result = cataloghttp.RepresentDemoStatus(value)
			}
			return readErr
		})
	return result, err
}

func (runtime *HTTPRuntime) PlanDemoRemoval(
	ctx context.Context,
	access cataloghttp.Access,
	installationID uuid.UUID,
) (cataloghttp.DemoRemovalPlan, error) {
	var result cataloghttp.DemoRemovalPlan
	err := runtime.demoSnapshot(ctx, access, cataloghttp.PermissionDemoManage,
		"catalog.installation.demo.removal_plan",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			value, planErr := runtime.lifecycle.PlanDemoRemoval(
				scoped, scope, installationID,
			)
			if planErr == nil {
				result = cataloghttp.RepresentDemoRemovalPlan(value)
			}
			return planErr
		})
	return result, err
}

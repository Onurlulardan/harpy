package iam

import (
	"context"

	"github.com/google/uuid"
)

func (runtime *FormsLifecycleAccessRuntime) tenantHeadersMatch(
	ctx context.Context, request FormsLifecycleAccessRequest, tenantID uuid.UUID,
) bool {
	if runtime == nil || runtime.shared == nil ||
		!optionalTenantIDMatches(request.TenantIDValues, tenantID) {
		return false
	}
	if len(request.TenantSlugValues) == 0 {
		return true
	}
	slug, valid := exactNonempty(request.TenantSlugValues, 100)
	if !valid {
		return false
	}
	resolved, err := runtime.shared.pool.ResolveActiveTenantIDBySlug(ctx, slug)
	return err == nil && resolved == tenantID
}

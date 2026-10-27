package iam

import (
	"context"

	"github.com/google/uuid"
)

func (runtime *PlatformAccessRuntime) headersMatch(
	ctx context.Context,
	request PlatformAccessRequest,
	tenantID uuid.UUID,
) bool {
	if !optionalTenantIDMatches(request.TenantIDValues, tenantID) {
		return false
	}
	if len(request.TenantSlugValues) == 0 {
		return true
	}
	slug, valid := exactNonempty(request.TenantSlugValues, 100)
	if !valid {
		return false
	}
	resolved, err := runtime.base.pool.ResolveActiveTenantIDBySlug(ctx, slug)
	return err == nil && resolved == tenantID
}

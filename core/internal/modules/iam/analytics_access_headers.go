package iam

import (
	"context"

	"github.com/google/uuid"
)

func (runtime *AnalyticsAccessRuntime) tenantHeadersMatch(
	ctx context.Context, request AnalyticsAccessRequest, tenantID uuid.UUID,
) bool {
	if runtime == nil || runtime.shared == nil {
		return false
	}
	return tenantAccessHeadersMatch(ctx, runtime.shared.pool,
		request.TenantIDValues, request.TenantSlugValues, tenantID)
}

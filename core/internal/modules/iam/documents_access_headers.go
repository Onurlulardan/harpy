package iam

import (
	"context"

	"github.com/google/uuid"
)

func (runtime *DocumentsAccessRuntime) tenantHeadersMatch(
	ctx context.Context, request DocumentsAccessRequest, tenantID uuid.UUID,
) bool {
	if runtime == nil || runtime.shared == nil {
		return false
	}
	return tenantAccessHeadersMatch(ctx, runtime.shared.pool,
		request.TenantIDValues, request.TenantSlugValues, tenantID)
}

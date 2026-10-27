package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func billingBearer(values []string) (string, bool) {
	value, valid := exactNonempty(values, accesstoken.MaximumWireSize+7)
	if !valid || !strings.HasPrefix(value, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(value, "Bearer ")
	return token, token != "" && len(token) <= accesstoken.MaximumWireSize &&
		!strings.ContainsAny(token, " \t\r\n")
}

func (runtime *BillingAccessRuntime) tenantHeadersMatch(
	ctx context.Context,
	request BillingAccessRequest,
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
	resolved, err := runtime.pool.ResolveActiveTenantIDBySlug(
		ctx, slug,
	)
	return err == nil && resolved == tenantID
}

package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func formsRecordBearer(values []string) (string, bool) {
	value, valid := exactNonempty(values, accesstoken.MaximumWireSize+7)
	if !valid || !strings.HasPrefix(value, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(value, "Bearer ")
	return token, token != "" && len(token) <= accesstoken.MaximumWireSize &&
		!strings.ContainsAny(token, " \t\r\n")
}

type tenantSlugResolver interface {
	ResolveActiveTenantIDBySlug(context.Context, string) (uuid.UUID, error)
}

func tenantAccessHeadersMatch(
	ctx context.Context,
	resolver tenantSlugResolver,
	tenantIDValues []string,
	tenantSlugValues []string,
	tenantID uuid.UUID,
) bool {
	if !optionalTenantIDMatches(tenantIDValues, tenantID) {
		return false
	}
	if len(tenantSlugValues) == 0 {
		return true
	}
	if resolver == nil {
		return false
	}
	slug, valid := exactNonempty(tenantSlugValues, 100)
	if !valid {
		return false
	}
	resolved, err := resolver.ResolveActiveTenantIDBySlug(ctx, slug)
	return err == nil && resolved == tenantID
}

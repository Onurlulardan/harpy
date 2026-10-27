package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *tenantOperations) readUsage(
	ctx context.Context,
	access billingapi.Access,
	authority generations,
) (billingapi.TenantUsageSnapshot, error) {
	var result billingapi.TenantUsageSnapshot
	err := application.inSnapshot(
		ctx, access, authority, "billing.usage.read",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			queries, bindErr := application.bindings.UsageQueries(scope)
			if bindErr != nil || queries == nil {
				return billingapi.ErrAuthorityUnavailable
			}
			value, queryErr := queries.Get(scoped, scope)
			if queryErr == nil {
				result = value
			}
			return queryErr
		},
	)
	if err != nil {
		return billingapi.TenantUsageSnapshot{}, normalizeUsageQueryError(err)
	}
	if !result.ValidForTenant(access.TenantID()) {
		return billingapi.TenantUsageSnapshot{}, billingapi.ErrUnavailable
	}
	return result, nil
}

package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *tenantOperations) readSubscription(
	ctx context.Context,
	access billingapi.Access,
	authority generations,
) (billingapi.TenantSubscriptionSnapshot, error) {
	var result billingapi.TenantSubscriptionSnapshot
	err := application.inSnapshot(
		ctx, access, authority, "billing.subscription.read",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			queries, bindErr := application.bindings.SubscriptionQueries(scope)
			if bindErr != nil || queries == nil {
				return billingapi.ErrAuthorityUnavailable
			}
			value, queryErr := queries.GetCurrent(scoped, scope)
			if queryErr == nil {
				result = value
			}
			return queryErr
		},
	)
	if err != nil {
		return billingapi.TenantSubscriptionSnapshot{}, normalizeSubscriptionQueryError(err)
	}
	if !result.ValidForTenant(access.TenantID()) {
		return billingapi.TenantSubscriptionSnapshot{}, billingapi.ErrUnavailable
	}
	return result, nil
}

package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantUsageQueries struct{ queries *usageapp.Queries }

func (adapter *tenantUsageQueries) Get(
	ctx context.Context,
	scope platformtx.TenantScope,
) (billingapi.TenantUsageSnapshot, error) {
	value, err := adapter.queries.Get(ctx, scope)
	if err != nil {
		return billingapi.TenantUsageSnapshot{}, err
	}
	return usageSnapshot(value)
}

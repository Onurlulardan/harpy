package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (application *tenantOperations) GetUsage(
	ctx context.Context,
	request billingapi.SubscriptionAccess,
) (billingapi.UsageView, error) {
	authority, err := application.resolve(
		ctx, request.Access, billingapi.PermissionBillingUsageRead,
	)
	if err != nil {
		return billingapi.UsageView{}, err
	}
	value, err := application.readUsage(ctx, request.Access, authority)
	if err != nil {
		return billingapi.UsageView{}, err
	}
	return value.View, nil
}

package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (application *tenantOperations) GetCurrentSubscription(
	ctx context.Context,
	request billingapi.SubscriptionAccess,
) (billingapi.SubscriptionView, error) {
	authority, err := application.resolve(
		ctx, request.Access, billingapi.PermissionBillingRead,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	value, err := application.readSubscription(ctx, request.Access, authority)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	return value.View, nil
}

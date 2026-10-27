package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantSubscriptionQueries struct{ queries *subscriptionapp.Queries }

func (adapter *tenantSubscriptionQueries) GetCurrent(
	ctx context.Context,
	scope platformtx.TenantScope,
) (billingapi.TenantSubscriptionSnapshot, error) {
	value, err := adapter.queries.GetCurrent(ctx, scope)
	if err != nil {
		return billingapi.TenantSubscriptionSnapshot{}, err
	}
	return subscriptionSnapshot(value)
}

type tenantSubscriptionCommands struct{ commands *subscriptionapp.Commands }

func (adapter *tenantSubscriptionCommands) ChangeAtPeriodEnd(
	ctx context.Context,
	request billingapi.TenantSubscriptionChange,
) error {
	_, err := adapter.commands.ChangeAtPeriodEnd(
		ctx, subscriptionapp.ChangeAtPeriodEndRequest{
			ActorID: request.ActorID, SubscriptionID: request.SubscriptionID,
			SourceOperationID:       request.SourceOperationID,
			TargetOfferingVersionID: request.TargetOfferingVersionID,
			TargetPriceVersionID:    request.TargetPriceVersionID,
			ExpectedRevision:        request.ExpectedRevision,
		},
	)
	return err
}

func (adapter *tenantSubscriptionCommands) CancelAtPeriodEnd(
	ctx context.Context,
	request billingapi.TenantSubscriptionCancellation,
) error {
	_, err := adapter.commands.CancelAtPeriodEnd(
		ctx, subscriptionapp.CancelAtPeriodEndRequest{
			ActorID: request.ActorID, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision, Reason: request.Reason,
		},
	)
	return err
}

func (adapter *tenantSubscriptionCommands) ResumeCancellation(
	ctx context.Context,
	request billingapi.TenantSubscriptionResume,
) error {
	_, err := adapter.commands.ResumeCancellation(
		ctx, subscriptionapp.ResumeCancellationRequest{
			ActorID: request.ActorID, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision,
		},
	)
	return err
}

func (adapter *tenantSubscriptionCommands) SetPlatformTransition(
	ctx context.Context,
	request subscriptionapp.PlatformTransitionRequest,
) (subscriptionapp.Result, error) {
	return adapter.commands.SetPlatformTransition(ctx, request)
}

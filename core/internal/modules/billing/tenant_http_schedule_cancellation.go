package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *tenantOperations) ScheduleCancellation(
	ctx context.Context,
	request billingapi.ScheduleCancellationCommand,
) (billingapi.SubscriptionView, error) {
	authority, err := application.resolve(
		ctx, request.Access, billingapi.PermissionBillingManage,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	if request.ExpectedVersion < 1 || !validReason(request.Reason) ||
		!validIdempotencyKey(request.IdempotencyKey) {
		return billingapi.SubscriptionView{}, billingapi.ErrInvalid
	}
	current, err := application.readSubscription(ctx, request.Access, authority)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	if err = subscriptionMatches(request.SubscriptionID, current); err != nil {
		return billingapi.SubscriptionView{}, err
	}
	sourceID, err := operationIdentity(
		request.Access, subscriptionapp.OperationCancelAtPeriodEnd, request.IdempotencyKey,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	err = application.inCommand(
		ctx, request.Access, authority, subscriptionapp.OperationCancelAtPeriodEnd,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			commands, bindErr := application.bindings.SubscriptionCommands(scope)
			if bindErr != nil || commands == nil {
				return billingapi.ErrAuthorityUnavailable
			}
			commandErr := commands.CancelAtPeriodEnd(
				scoped, billingapi.TenantSubscriptionCancellation{
					ActorID: request.Access.PrincipalID(), SubscriptionID: current.ID,
					SourceOperationID: sourceID, ExpectedRevision: request.ExpectedVersion,
					Reason: request.Reason,
				},
			)
			return commandErr
		},
	)
	if err != nil {
		return billingapi.SubscriptionView{}, normalizeSubscriptionCommandError(err)
	}
	current, err = application.readSubscription(ctx, request.Access, authority)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	return current.View, nil
}

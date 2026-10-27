package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (reports *platformReports) TransitionSubscription(
	ctx context.Context,
	command billingapi.TransitionSubscriptionCommand,
) (billingapi.SubscriptionView, error) {
	if reports == nil || reports.global == nil || ctx == nil || ctx.Err() != nil ||
		!command.Access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		command.Access.Scope() != billingapi.AccessScopeGlobal {
		return billingapi.SubscriptionView{}, billingapi.ErrForbidden
	}
	if command.ExpectedVersion < 1 || !validReason(command.Reason) {
		return billingapi.SubscriptionView{}, billingapi.ErrInvalid
	}
	subscriptionID, valid := parseV7(command.SubscriptionID)
	operationID, operationErr := platformReportOperationID(
		command.Access, subscriptionID, command.IdempotencyKey,
	)
	if !valid || !validPlatformSubscriptionState(command.TargetState) || operationErr != nil {
		return billingapi.SubscriptionView{}, billingapi.ErrInvalid
	}
	target, err := reports.resolveTransitionTarget(ctx, command.Access, subscriptionID)
	if err != nil {
		return billingapi.SubscriptionView{}, normalizePlatformTransitionError(err)
	}
	result, err := reports.executePlatformTransition(
		ctx, command, target, subscriptionID, operationID,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, normalizePlatformTransitionError(err)
	}
	return platformSubscriptionView(result.Subscription, target.Provider), nil
}

func (reports *platformReports) resolveTransitionTarget(
	ctx context.Context,
	access billingapi.Access,
	subscriptionID uuid.UUID,
) (platformreportapp.TransitionTarget, error) {
	capability, found := access.GlobalCommandCapability()
	if !found {
		return platformreportapp.TransitionTarget{}, billingapi.ErrAuthorityUnavailable
	}
	var target platformreportapp.TransitionTarget
	err := reports.global.Command(ctx, capability, func(
		scoped context.Context, scope platformtx.GlobalScope, authority platformtx.OwnerAuthority,
	) error {
		if scope.OperationID() != operationSubscriptionTransition {
			return billingapi.ErrAuthorityUnavailable
		}
		var resolveErr error
		target, resolveErr = reports.locator.ResolveTransitionTarget(
			scoped, scope, authority, subscriptionID,
		)
		return resolveErr
	})
	return target, err
}

func (reports *platformReports) executePlatformTransition(
	ctx context.Context,
	command billingapi.TransitionSubscriptionCommand,
	target platformreportapp.TransitionTarget,
	subscriptionID, operationID uuid.UUID,
) (subscriptionapp.Result, error) {
	request, err := platformtx.AdmitCommandScope(
		command.Access.PrincipalID(), target.TenantID,
		subscriptionapp.OperationPlatformTransition, billingAuthorityGroup,
		target.GroupGeneration, map[string]int64{
			"billing": target.BillingGeneration, "forms": target.FormsGeneration,
		},
	)
	if err != nil {
		return subscriptionapp.Result{}, billingapi.ErrAuthorityUnavailable
	}
	var result subscriptionapp.Result
	err = reports.coordinator.InTenant(ctx, request, func(
		scoped context.Context,
		scope platformtx.TenantScope,
	) error {
		bound, bindErr := reports.bindings.SubscriptionCommands(scope)
		if bindErr != nil {
			return billingapi.ErrAuthorityUnavailable
		}
		commands, valid := bound.(platformSubscriptionCommands)
		if !valid {
			return billingapi.ErrAuthorityUnavailable
		}
		var executeErr error
		result, executeErr = commands.SetPlatformTransition(scoped, subscriptionapp.PlatformTransitionRequest{
			ActorID: command.Access.PrincipalID(), SubscriptionID: subscriptionID,
			SourceOperationID: operationID, ExpectedRevision: command.ExpectedVersion,
			TargetState: command.TargetState, Reason: command.Reason,
		})
		return executeErr
	})
	return result, err
}

func validPlatformSubscriptionState(value string) bool {
	switch value {
	case "pending", "trialing", "active", "past_due", "paused", "cancelled", "expired":
		return true
	default:
		return false
	}
}

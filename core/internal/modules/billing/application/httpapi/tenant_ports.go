package httpapi

import "context"

type TenantOperations interface {
	GetCurrentSubscription(context.Context, SubscriptionAccess) (SubscriptionView, error)
	GetUsage(context.Context, SubscriptionAccess) (UsageView, error)
	ScheduleChange(context.Context, ScheduleChangeCommand) (SubscriptionView, error)
	ScheduleCancellation(context.Context, ScheduleCancellationCommand) (SubscriptionView, error)
	ResumeCancellation(context.Context, ResumeCancellationCommand) (SubscriptionView, error)
	RedeemDiscount(context.Context, RedeemDiscountCommand) (SubscriptionView, error)
	ValidateDiscount(context.Context, ValidateDiscountCommand) (DiscountView, error)
}

type TenantRuntime interface {
	TenantOperations
	LegacyIdentityResolver
}

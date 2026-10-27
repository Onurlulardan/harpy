package httpapi

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type TenantGenerationResolver interface {
	ResolveTenantBillingGenerations(
		context.Context, uuid.UUID,
	) (group, billing int64, err error)
}

type TenantScopeCoordinator interface {
	InTenant(
		context.Context,
		platformtx.ScopeRequest,
		func(context.Context, platformtx.TenantScope) error,
	) error
}

type TenantSubscriptionQueryPort interface {
	GetCurrent(
		context.Context, platformtx.TenantScope,
	) (TenantSubscriptionSnapshot, error)
}

type TenantUsageQueryPort interface {
	Get(context.Context, platformtx.TenantScope) (TenantUsageSnapshot, error)
}

type TenantDiscountQueryPort interface {
	Validate(
		context.Context, platformtx.TenantScope, TenantDiscountLookup,
	) (TenantDiscountSnapshot, error)
	ResolveCurrent(
		context.Context, platformtx.TenantScope, string,
	) (TenantDiscountSnapshot, error)
}

type TenantSubscriptionCommandPort interface {
	ChangeAtPeriodEnd(context.Context, TenantSubscriptionChange) error
	CancelAtPeriodEnd(context.Context, TenantSubscriptionCancellation) error
	ResumeCancellation(context.Context, TenantSubscriptionResume) error
}

type TenantDiscountCommandPort interface {
	Redeem(context.Context, TenantDiscountRedemption) error
}

type TenantOperationBindings interface {
	SubscriptionQueries(platformtx.TenantScope) (TenantSubscriptionQueryPort, error)
	UsageQueries(platformtx.TenantScope) (TenantUsageQueryPort, error)
	DiscountQueries(platformtx.TenantScope) (TenantDiscountQueryPort, error)
	SubscriptionCommands(platformtx.TenantScope) (TenantSubscriptionCommandPort, error)
	DiscountCommands(platformtx.TenantScope) (TenantDiscountCommandPort, error)
}

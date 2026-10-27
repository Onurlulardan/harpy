package billing

import (
	"context"

	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	billingreports "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/reports"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PlatformReportQueries interface {
	ListSubscriptions(context.Context, platformtx.GlobalScope, billingreports.QueryRequest) (billingreports.RecordResult, error)
	ReadAuditEvents(context.Context, platformtx.GlobalScope, billingreports.QueryRequest) (billingreports.RecordResult, error)
	ExecuteLegacyEvents(context.Context, platformtx.GlobalScope, billingreports.QueryRequest) (billingreports.RecordResult, error)
	GetRevenue(context.Context, platformtx.GlobalScope, billingreports.QueryRequest) (billingreports.RevenueResult, error)
}

type PlatformTransitionResolver interface {
	ResolveTransitionTarget(
		context.Context, platformtx.GlobalScope, platformtx.OwnerAuthority, uuid.UUID,
	) (platformreportapp.TransitionTarget, error)
}

type platformSubscriptionCommands interface {
	SetPlatformTransition(
		context.Context, subscriptionapp.PlatformTransitionRequest,
	) (subscriptionapp.Result, error)
}

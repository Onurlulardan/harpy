package billing

import (
	"context"
	"time"

	resourcepostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/resourceentitlement"
	resourceusage "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/resourceusage"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type MeteredResourceEntitlements struct {
	catalogs platformasync.WriteCatalogs
	nextID   usageapp.IDGenerator
	clock    usageapp.Clock
}

var _ billingapi.MeteredResourceMutationDecider = (*MeteredResourceEntitlements)(nil)

func NewMeteredResourceEntitlements(
	catalogs platformasync.WriteCatalogs,
	nextID usageapp.IDGenerator,
	clock func() time.Time,
) (*MeteredResourceEntitlements, error) {
	if _, _, err := catalogs.Snapshot(); err != nil || nextID == nil || clock == nil {
		return nil, resourceapp.ErrConfiguration
	}
	return &MeteredResourceEntitlements{
		catalogs: catalogs,
		nextID:   nextID,
		clock:    clock,
	}, nil
}

func (service *MeteredResourceEntitlements) DecideMeteredResourceMutation(
	ctx context.Context,
	scope platformtx.TenantScope,
	request billingapi.MeteredResourceMutationRequest,
) (billingapi.MeteredResourceMutationDecision, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || !request.Valid() ||
		request.TenantID != scope.TenantID() || request.ActorID != scope.ActorID() {
		return billingapi.MeteredResourceMutationDecision{}, resourceapp.ErrRequest
	}
	resourceStore, err := resourcepostgres.Bind(scope)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	usageStore, err := BindUsageStore(scope, service.catalogs)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	usageCommands, err := usageapp.NewCommands(usageStore, service.nextID, service.clock)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	resourceUsage, err := resourceusage.Bind(usageCommands)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	commands, err := resourceapp.NewCommands(
		resourceStore,
		resourceUsage,
		resourceapp.Clock(service.clock),
	)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	return commands.Decide(ctx, request)
}

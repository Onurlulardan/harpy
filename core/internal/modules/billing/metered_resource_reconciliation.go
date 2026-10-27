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

type MeteredResourceReconciliation struct {
	catalogs platformasync.WriteCatalogs
	nextID   usageapp.IDGenerator
	clock    usageapp.Clock
	verifier billingapi.MeteredResourceEvidenceVerifier
}

var _ billingapi.MeteredResourceReconciler = (*MeteredResourceReconciliation)(nil)

func NewMeteredResourceReconciliation(
	catalogs platformasync.WriteCatalogs,
	nextID usageapp.IDGenerator,
	clock func() time.Time,
	verifier billingapi.MeteredResourceEvidenceVerifier,
) (*MeteredResourceReconciliation, error) {
	if _, _, err := catalogs.Snapshot(); err != nil || nextID == nil ||
		clock == nil || verifier == nil {
		return nil, resourceapp.ErrConfiguration
	}
	return &MeteredResourceReconciliation{
		catalogs: catalogs, nextID: nextID, clock: clock, verifier: verifier,
	}, nil
}

func (service *MeteredResourceReconciliation) ReconcileMeteredResource(
	ctx context.Context,
	scope platformtx.TenantScope,
	request billingapi.MeteredResourceReconcileRequest,
) (billingapi.MeteredResourceReconcileDecision, error) {
	if service == nil || ctx == nil || ctx.Err() != nil || !request.Valid() ||
		request.TenantID != scope.TenantID() || request.ActorID != scope.ActorID() {
		return billingapi.MeteredResourceReconcileDecision{}, resourceapp.ErrRequest
	}
	query := billingapi.MeteredResourceEvidenceQuery{
		Resource: request.Resource, TenantID: request.TenantID,
		OwnerFactID: request.OwnerFactID,
	}
	evidence, err := service.verifier.VerifyMeteredResourceEvidence(ctx, scope, query)
	if err != nil || !evidence.ValidFor(query) {
		return billingapi.MeteredResourceReconcileDecision{}, resourceapp.ErrEvidence
	}
	resourceStore, err := resourcepostgres.Bind(scope)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	usageStore, err := BindUsageStore(scope, service.catalogs)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	usageCommands, err := usageapp.NewCommands(usageStore, service.nextID, service.clock)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	resourceUsage, err := resourceusage.Bind(usageCommands)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	commands, err := resourceapp.NewReconcileCommands(
		resourceStore,
		resourceUsage,
		resourceapp.Clock(service.clock),
	)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	return commands.Reconcile(ctx, request, evidence)
}

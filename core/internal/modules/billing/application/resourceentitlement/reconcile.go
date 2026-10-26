package resourceentitlement

import (
	"context"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

func (commands *ReconcileCommands) Reconcile(
	ctx context.Context,
	request billingapi.MeteredResourceReconcileRequest,
	evidence billingapi.MeteredResourceEvidence,
) (billingapi.MeteredResourceReconcileDecision, error) {
	query := billingapi.MeteredResourceEvidenceQuery{
		Resource: request.Resource, TenantID: request.TenantID,
		OwnerFactID: request.OwnerFactID,
	}
	if commands == nil || ctx == nil || ctx.Err() != nil || !request.Valid() ||
		!evidence.ValidFor(query) {
		return billingapi.MeteredResourceReconcileDecision{}, ErrEvidence
	}
	now := commands.clock()
	if !validInstant(now) {
		return billingapi.MeteredResourceReconcileDecision{}, ErrConfiguration
	}
	inspection, err := commands.store.InspectReconciliation(ctx, request)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{},
			fmt.Errorf("inspect billing metered resource reconciliation: %w", err)
	}
	if inspection.Stored != nil {
		return replayReconciliation(request, evidence, *inspection.Stored)
	}
	if !validReconciliationProjection(inspection.Projection, request) {
		return billingapi.MeteredResourceReconcileDecision{}, ErrUnavailable
	}
	decision, change, link, err := commands.planReconciliation(
		ctx, request, evidence, inspection.Projection, now,
	)
	if err != nil {
		return billingapi.MeteredResourceReconcileDecision{}, err
	}
	if !decision.ValidFor(request, evidence) {
		return billingapi.MeteredResourceReconcileDecision{}, ErrUnavailable
	}
	if err := commands.store.RecordReconciliation(ctx, ReconciliationRecord{
		Decision: decision, Evidence: evidence, Change: change,
		UsageDecisionID: link.DecisionID, UsageEntryID: link.EntryID,
		RecordedAt: now,
	}); err != nil {
		return billingapi.MeteredResourceReconcileDecision{},
			fmt.Errorf("record billing metered resource reconciliation: %w", err)
	}
	return decision, nil
}

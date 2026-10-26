package resourceentitlement

import (
	"context"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

func (commands *Commands) Decide(
	ctx context.Context,
	request billingapi.MeteredResourceMutationRequest,
) (billingapi.MeteredResourceMutationDecision, error) {
	if commands == nil || ctx == nil || ctx.Err() != nil || !request.Valid() {
		return billingapi.MeteredResourceMutationDecision{}, ErrRequest
	}
	now := commands.clock()
	if !validInstant(now) {
		return billingapi.MeteredResourceMutationDecision{}, ErrConfiguration
	}
	inspection, err := commands.store.Inspect(ctx, request)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{},
			fmt.Errorf("inspect billing metered resource: %w", err)
	}
	if inspection.Stored != nil {
		return replayDecision(request, *inspection.Stored)
	}
	if !validProjection(inspection.Projection, request) {
		return billingapi.MeteredResourceMutationDecision{}, ErrUnavailable
	}
	spec, err := resolveSpec(request.Resource)
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	decision := newDecision(request)
	var change *ProjectionChange
	var usageLink UsageLink
	switch request.Mutation {
	case billingapi.MeteredResourceUpdate, billingapi.MeteredResourcePublish,
		billingapi.MeteredResourceRevalidate:
		decision.FeatureEnabled, err = commands.store.ResolveFeature(ctx, spec.feature)
		decision.Allowed = decision.FeatureEnabled
	case billingapi.MeteredResourceCreate, billingapi.MeteredResourceRestore:
		decision, change, usageLink, err = commands.admitIncrease(
			ctx, request, spec, inspection.Projection, now,
		)
	case billingapi.MeteredResourceDelete:
		decision, change, usageLink, err = commands.release(
			ctx, request, spec, inspection.Projection, now,
		)
	default:
		err = ErrRequest
	}
	if err != nil {
		return billingapi.MeteredResourceMutationDecision{}, err
	}
	if !decision.ValidFor(request) {
		return billingapi.MeteredResourceMutationDecision{}, ErrUnavailable
	}
	if err := commands.store.Record(ctx, MutationRecord{
		Decision: decision, Change: change,
		UsageDecisionID: usageLink.DecisionID, UsageEntryID: usageLink.EntryID,
		RecordedAt: now,
	}); err != nil {
		return billingapi.MeteredResourceMutationDecision{},
			fmt.Errorf("record billing metered resource decision: %w", err)
	}
	return decision, nil
}

func newDecision(
	request billingapi.MeteredResourceMutationRequest,
) billingapi.MeteredResourceMutationDecision {
	return billingapi.MeteredResourceMutationDecision{
		Resource: request.Resource, Mutation: request.Mutation,
		TenantID: request.TenantID, ActorID: request.ActorID,
		OperationID:    request.ProposedOperationID,
		OwnerFactID:    request.ProposedOwnerFactID,
		IdempotencyKey: request.IdempotencyKey,
		RequestSHA256:  request.RequestSHA256,
	}
}

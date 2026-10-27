package publication

import (
	"context"
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) decideMutation(
	ctx context.Context,
	scope platformtx.TenantScope,
	identity ScopeIdentity,
	resourceKind ResourceKind,
	operationID, releaseID uuid.UUID,
	idempotencyKey string,
	requestSHA256 [32]byte,
) (billingapi.MeteredResourceMutationDecision, error) {
	resource, ok := entitlementResource(resourceKind)
	if !ok || nilPort(application.entitlements) {
		return billingapi.MeteredResourceMutationDecision{}, ErrUnavailable
	}
	request := billingapi.MeteredResourceMutationRequest{
		Resource: resource, Mutation: billingapi.MeteredResourcePublish,
		TenantID: identity.TenantID, ActorID: identity.ActorID,
		ProposedOperationID: operationID, ProposedOwnerFactID: releaseID,
		IdempotencyKey: idempotencyKey, RequestSHA256: requestSHA256,
	}
	decision, err := application.entitlements.DecideMeteredResourceMutation(ctx, scope, request)
	if errors.Is(err, billingapi.ErrMeteredResourceRequest) {
		return billingapi.MeteredResourceMutationDecision{}, ErrInvalid
	}
	if errors.Is(err, billingapi.ErrMeteredResourceIdempotencyReused) ||
		errors.Is(err, billingapi.ErrMeteredResourceStateConflict) {
		return billingapi.MeteredResourceMutationDecision{}, ErrConflict
	}
	if err != nil || !decision.ValidFor(request) {
		return billingapi.MeteredResourceMutationDecision{}, ErrUnavailable
	}
	if !decision.Allowed {
		return billingapi.MeteredResourceMutationDecision{}, ErrEntitlement
	}
	return decision, nil
}

func entitlementResource(kind ResourceKind) (billingapi.MeteredResourceKind, bool) {
	switch kind {
	case ResourceQuery:
		return billingapi.MeteredResourceAPIEndpoint, true
	case ResourceDashboard:
		return billingapi.MeteredResourceAnalyticsDashboard, true
	default:
		return "", false
	}
}

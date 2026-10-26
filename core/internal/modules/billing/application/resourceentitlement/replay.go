package resourceentitlement

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

func replayDecision(
	request billingapi.MeteredResourceMutationRequest,
	stored StoredMutation,
) (billingapi.MeteredResourceMutationDecision, error) {
	decision := stored.Decision
	if !validInstant(stored.RecordedAt) || decision.Replay ||
		decision.Resource != request.Resource || decision.Mutation != request.Mutation ||
		decision.TenantID != request.TenantID || decision.ActorID != request.ActorID ||
		decision.IdempotencyKey != request.IdempotencyKey ||
		decision.RequestSHA256 != request.RequestSHA256 ||
		request.Mutation != billingapi.MeteredResourceCreate &&
			decision.OwnerFactID != request.ProposedOwnerFactID {
		return billingapi.MeteredResourceMutationDecision{}, ErrIdempotencyReused
	}
	decision.Replay = true
	if !decision.ValidFor(request) {
		return billingapi.MeteredResourceMutationDecision{}, ErrUnavailable
	}
	return decision, nil
}

func validProjection(
	projection *ResourceProjection,
	request billingapi.MeteredResourceMutationRequest,
) bool {
	if projection == nil {
		return true
	}
	return projection.Resource == request.Resource &&
		projection.OwnerFactID == request.ProposedOwnerFactID &&
		projection.State.Valid() && projection.Revision > 0 &&
		validInstant(projection.UpdatedAt)
}

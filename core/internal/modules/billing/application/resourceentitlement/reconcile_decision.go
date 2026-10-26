package resourceentitlement

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"

func newReconciliationDecision(
	request billingapi.MeteredResourceReconcileRequest,
	evidence billingapi.MeteredResourceEvidence,
) billingapi.MeteredResourceReconcileDecision {
	return billingapi.MeteredResourceReconcileDecision{
		Resource: request.Resource, TenantID: request.TenantID, ActorID: request.ActorID,
		OperationID: request.ProposedOperationID, OwnerFactID: request.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: request.RequestSHA256,
		EvidenceSHA256: evidence.EvidenceSHA256, Exists: evidence.Exists,
		Active: evidence.Active, IncludedInOpening: evidence.IncludedInOpening,
	}
}

func replayReconciliation(
	request billingapi.MeteredResourceReconcileRequest,
	evidence billingapi.MeteredResourceEvidence,
	stored StoredReconciliation,
) (billingapi.MeteredResourceReconcileDecision, error) {
	decision := stored.Decision
	if !validInstant(stored.RecordedAt) || decision.Replay ||
		decision.Resource != request.Resource || decision.TenantID != request.TenantID ||
		decision.ActorID != request.ActorID || decision.OwnerFactID != request.OwnerFactID ||
		decision.IdempotencyKey != request.IdempotencyKey ||
		decision.RequestSHA256 != request.RequestSHA256 ||
		decision.EvidenceSHA256 != evidence.EvidenceSHA256 ||
		decision.Exists != evidence.Exists || decision.Active != evidence.Active ||
		decision.IncludedInOpening != evidence.IncludedInOpening {
		return billingapi.MeteredResourceReconcileDecision{}, ErrIdempotencyReused
	}
	decision.Replay = true
	if !decision.ValidFor(request, evidence) {
		return billingapi.MeteredResourceReconcileDecision{}, ErrUnavailable
	}
	return decision, nil
}

func validReconciliationProjection(
	projection *ResourceProjection,
	request billingapi.MeteredResourceReconcileRequest,
) bool {
	if projection == nil {
		return true
	}
	return projection.Resource == request.Resource &&
		projection.OwnerFactID == request.OwnerFactID && projection.State.Valid() &&
		projection.Revision > 0 && validInstant(projection.UpdatedAt)
}

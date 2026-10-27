package billingapi

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type MeteredResourceEvidenceQuery struct {
	Resource    MeteredResourceKind
	TenantID    [16]byte
	OwnerFactID [16]byte
}

func (query MeteredResourceEvidenceQuery) Valid() bool {
	return query.Resource.Valid() && query.TenantID != ([16]byte{}) &&
		validUUIDv7(query.OwnerFactID)
}

type MeteredResourceEvidence struct {
	Resource          MeteredResourceKind
	TenantID          [16]byte
	OwnerFactID       [16]byte
	Exists            bool
	Active            bool
	IncludedInOpening bool
	EvidenceSHA256    [32]byte
	MeasuredAt        time.Time
}

func (evidence MeteredResourceEvidence) ValidFor(
	query MeteredResourceEvidenceQuery,
) bool {
	return query.Valid() && evidence.Resource == query.Resource &&
		evidence.TenantID == query.TenantID &&
		evidence.OwnerFactID == query.OwnerFactID &&
		evidence.EvidenceSHA256 != ([32]byte{}) && validEvidenceTime(evidence.MeasuredAt) &&
		(evidence.Exists || !evidence.Active) &&
		(!evidence.IncludedInOpening || evidence.Exists && evidence.Active)
}

type MeteredResourceEvidenceVerifier interface {
	VerifyMeteredResourceEvidence(
		context.Context,
		platformtx.TenantScope,
		MeteredResourceEvidenceQuery,
	) (MeteredResourceEvidence, error)
}

type MeteredResourceReconcileRequest struct {
	Resource            MeteredResourceKind
	TenantID            [16]byte
	ActorID             [16]byte
	ProposedOperationID [16]byte
	OwnerFactID         [16]byte
	IdempotencyKey      string
	RequestSHA256       [32]byte
}

func (request MeteredResourceReconcileRequest) Valid() bool {
	return request.Resource.Valid() && request.TenantID != ([16]byte{}) &&
		request.ActorID != ([16]byte{}) && validUUIDv7(request.ProposedOperationID) &&
		validUUIDv7(request.OwnerFactID) && validIdempotencyKey(request.IdempotencyKey) &&
		request.RequestSHA256 != ([32]byte{})
}

type MeteredResourceReconcileDecision struct {
	Resource          MeteredResourceKind
	TenantID          [16]byte
	ActorID           [16]byte
	OperationID       [16]byte
	OwnerFactID       [16]byte
	IdempotencyKey    string
	RequestSHA256     [32]byte
	EvidenceSHA256    [32]byte
	Exists            bool
	Active            bool
	IncludedInOpening bool
	Changed           bool
	Replay            bool
	GaugeDelta        int64
	Current           int64
	Limit             int64
	Unlimited         bool
}

func (decision MeteredResourceReconcileDecision) ValidFor(
	request MeteredResourceReconcileRequest,
	evidence MeteredResourceEvidence,
) bool {
	query := MeteredResourceEvidenceQuery{
		Resource: request.Resource, TenantID: request.TenantID,
		OwnerFactID: request.OwnerFactID,
	}
	if !request.Valid() || !evidence.ValidFor(query) ||
		decision.Resource != request.Resource ||
		decision.TenantID != request.TenantID || decision.ActorID != request.ActorID ||
		decision.OwnerFactID != request.OwnerFactID ||
		decision.IdempotencyKey != request.IdempotencyKey ||
		decision.RequestSHA256 != request.RequestSHA256 ||
		decision.EvidenceSHA256 != evidence.EvidenceSHA256 ||
		decision.Exists != evidence.Exists || decision.Active != evidence.Active ||
		decision.IncludedInOpening != evidence.IncludedInOpening ||
		!validUUIDv7(decision.OperationID) || decision.Current < 0 ||
		decision.Limit < 0 || decision.Unlimited && decision.Limit != 0 ||
		decision.GaugeDelta < -1 || decision.GaugeDelta > 1 {
		return false
	}
	if !decision.Replay && decision.OperationID != request.ProposedOperationID {
		return false
	}
	return decision.Changed || decision.GaugeDelta == 0
}

type MeteredResourceReconciler interface {
	ReconcileMeteredResource(
		context.Context,
		platformtx.TenantScope,
		MeteredResourceReconcileRequest,
	) (MeteredResourceReconcileDecision, error)
}

func validEvidenceTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

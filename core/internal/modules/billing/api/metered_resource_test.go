package billingapi

import (
	"errors"
	"testing"
	"time"
)

func TestMeteredResourceMutationContractRejectsMalformedIdentity(t *testing.T) {
	request := MeteredResourceMutationRequest{
		Resource: MeteredResourceAnalyticsDashboard, Mutation: MeteredResourceCreate,
		TenantID: apiTestV7(1), ActorID: apiTestV7(2),
		ProposedOperationID: apiTestV7(3),
		ProposedOwnerFactID: apiTestV7(4),
		IdempotencyKey:      "dashboard:create:1", RequestSHA256: [32]byte{1},
	}
	if !request.Valid() {
		t.Fatal("valid request rejected")
	}
	request.ProposedOwnerFactID = [16]byte{1}
	if request.Valid() {
		t.Fatal("non-v7 owner fact accepted")
	}
}

func TestMeteredResourceDeleteOutcomeIsBuilderIndependent(t *testing.T) {
	request := MeteredResourceMutationRequest{
		Resource: MeteredResourceAPIEndpoint, Mutation: MeteredResourceDelete,
		TenantID: apiTestV7(1), ActorID: apiTestV7(2),
		ProposedOperationID: apiTestV7(3),
		ProposedOwnerFactID: apiTestV7(4),
		IdempotencyKey:      "endpoint-delete-1", RequestSHA256: [32]byte{1},
	}
	decision := MeteredResourceMutationDecision{
		Resource: request.Resource, Mutation: request.Mutation,
		TenantID: request.TenantID, ActorID: request.ActorID,
		OperationID: request.ProposedOperationID, OwnerFactID: request.ProposedOwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: request.RequestSHA256,
		Allowed: true, GaugeDelta: -1, Current: 1, Limit: 0, Unlimited: true,
	}
	if !decision.ValidFor(request) {
		t.Fatal("builder-independent delete decision rejected")
	}
}

func TestReconcileEvidenceRequiresExactOwnerAndUTCMicroseconds(t *testing.T) {
	query := MeteredResourceEvidenceQuery{
		Resource: MeteredResourceAPIEndpoint, TenantID: apiTestV7(1),
		OwnerFactID: apiTestV7(2),
	}
	evidence := MeteredResourceEvidence{
		Resource: query.Resource, TenantID: query.TenantID, OwnerFactID: query.OwnerFactID,
		Exists: true, Active: true, IncludedInOpening: true,
		EvidenceSHA256: [32]byte{1},
		MeasuredAt:     time.Date(2026, 8, 2, 10, 0, 0, 123000, time.UTC),
	}
	if !evidence.ValidFor(query) {
		t.Fatal("valid authoritative evidence rejected")
	}
	evidence.OwnerFactID = apiTestV7(3)
	if evidence.ValidFor(query) {
		t.Fatal("cross-owner evidence accepted")
	}
}

func apiTestV7(seed byte) [16]byte {
	value := [16]byte{seed}
	value[6] = 0x70
	value[8] = 0x80
	return value
}

func TestMeteredResourceErrorsArePublicSentinels(t *testing.T) {
	wrapped := errors.New("wrapper: " + ErrMeteredResourceUnavailable.Error())
	if errors.Is(wrapped, ErrMeteredResourceUnavailable) {
		t.Fatal("text equality must not classify errors")
	}
	if !errors.Is(ErrMeteredResourceUnavailable, ErrMeteredResourceUnavailable) {
		t.Fatal("public sentinel cannot be classified")
	}
}

package customerlifecycle

import (
	"context"
	"errors"
	"testing"
)

func TestOffboardRequiresPurposeBoundStepUp(t *testing.T) {
	service, access, store, readiness := newFixtureService(t)
	operation, err := service.RequestOffboard(context.Background(), AsyncCommand{
		TenantID: fixtureTargetID, ExpectedRevision: 4,
		IdempotencyKey: "tenant-offboard-key-01",
	})
	if err != nil || operation.ID == [16]byte{} {
		t.Fatalf("operation=%+v err=%v", operation, err)
	}
	request := access.requests[0]
	if len(access.requests) != 1 || request.Capability != CapabilityOffboard ||
		!request.RequireStepUp || request.StepUpPurpose == [32]byte{} || readiness.calls != 0 {
		t.Fatalf("access=%+v readiness=%d", request, readiness.calls)
	}
	if store.asyncRequest.Kind != OperationOffboard || store.asyncRequest.Proof != (PurgeProof{}) {
		t.Fatalf("store request=%+v", store.asyncRequest)
	}
}

func TestPurgeUsesTrustedReadinessProof(t *testing.T) {
	service, _, store, readiness := newFixtureService(t)
	_, err := service.RequestPurge(context.Background(), AsyncCommand{
		TenantID: fixtureTargetID, ExpectedRevision: 9,
		IdempotencyKey: "tenant-purge-key-0001",
	})
	if err != nil || readiness.calls != 1 || store.asyncRequest.Proof != readiness.proof {
		t.Fatalf("err=%v calls=%d proof=%+v", err, readiness.calls, store.asyncRequest.Proof)
	}
}

func TestPurgeRejectsMissingTrustedProofBeforeStorage(t *testing.T) {
	service, _, store, readiness := newFixtureService(t)
	readiness.proof = PurgeProof{}
	_, err := service.RequestPurge(context.Background(), AsyncCommand{
		TenantID: fixtureTargetID, ExpectedRevision: 9,
		IdempotencyKey: "tenant-purge-key-0002",
	})
	if !errors.Is(err, ErrPurgeProof) || store.asyncRequest.TargetTenantID != [16]byte{} {
		t.Fatalf("err=%v request=%+v", err, store.asyncRequest)
	}
}

func TestStaleStepUpIsRejected(t *testing.T) {
	service, access, _, _ := newFixtureService(t)
	access.now = fixtureNow.Add(-20 * 60 * 1e9)
	_, err := service.RequestOffboard(context.Background(), AsyncCommand{
		TenantID: fixtureTargetID, ExpectedRevision: 1,
		IdempotencyKey: "tenant-offboard-key-02",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err=%v", err)
	}
}

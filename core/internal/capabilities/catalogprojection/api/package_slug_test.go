package catalogprojection

import (
	"crypto/sha256"
	"testing"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	"github.com/google/uuid"
)

const (
	testPackageID = "e3000000-0000-0000-0000-000000000001"
	testReleaseID = "019c0000-0000-7000-8000-000000000001"
	testComponent = "019c0000-0000-7000-8000-000000000002"
	testReceipt   = "019c0000-0000-7000-8000-000000000003"
)

type decoderStub struct{ materialization Materialization }

func (stub decoderStub) Decode(
	_ []byte,
	_ contract.Component,
) (Materialization, error) {
	return stub.materialization.Clone(), nil
}

func TestVerifyBindsPackageSlugAndAcceptsStablePackageID(t *testing.T) {
	document := []byte(`{"contract":"fixture.v1"}`)
	digest := sha256.Sum256(document)
	materialization := Materialization{
		Kind: "FIXTURE", ModuleSlug: "human-resources",
		Identity: "fixture", Document: document,
		SemanticSHA256: digest, ItemCount: 1,
	}
	application := &Application{
		owner: "workflow", contract: "fixture.v1",
		decoder: decoderStub{materialization: materialization},
	}
	request := fixtureRequest(document, "human-resources")

	verified, err := application.verify(request)
	if err != nil || verified.packageID.String() != testPackageID {
		t.Fatalf("verify stable package id: %v", err)
	}

	request.PackageSlug = "sales-crm"
	if _, err = application.verify(request); err != ErrInvalid {
		t.Fatalf("expected package/module slug mismatch, got %v", err)
	}
}

func TestSnapshotReplayRejectsPackageSlugDrift(t *testing.T) {
	document := []byte(`{"contract":"fixture.v1"}`)
	digest := sha256.Sum256(document)
	request := fixtureRequest(document, "human-resources")
	verified := verifiedRequest{
		request: request, packageID: uuid.MustParse(testPackageID),
		releaseID: uuid.MustParse(testReleaseID),
		materialization: Materialization{
			Kind: "FIXTURE", ModuleSlug: request.PackageSlug,
			Identity: "fixture", Document: document,
			SemanticSHA256: digest, ItemCount: 1,
		},
	}
	requestSHA := sha256.Sum256([]byte("request"))
	snapshot := fixtureSnapshot(verified, requestSHA)
	if !snapshotMatches(snapshot, verified, requestSHA) {
		t.Fatal("fixture snapshot should match")
	}
	snapshot.PackageSlug = "sales-crm"
	if snapshotMatches(snapshot, verified, requestSHA) {
		t.Fatal("package slug drift must conflict")
	}
}

func TestReceiptEchoesPackageSlug(t *testing.T) {
	snapshot := Snapshot{
		ReceiptID:   uuid.MustParse(testReceipt),
		PackageID:   uuid.MustParse(testPackageID),
		PackageSlug: "human-resources", PackageVersion: "1.0.0",
		ReleaseID: uuid.MustParse(testReleaseID),
		Component: contract.Component{ID: testComponent},
		State:     contract.ProvisionAvailable,
		ReadyAt:   time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
	}
	receipt := receiptFromSnapshot(snapshot)
	if receipt.PackageSlug != snapshot.PackageSlug || !receipt.Ready ||
		receipt.CompletedAt != snapshot.ReadyAt {
		t.Fatalf("receipt did not echo ready snapshot: %#v", receipt)
	}
}

func TestPendingReceiptHasRetryWithoutClaimingReady(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	snapshot := Snapshot{
		ReceiptID: uuid.MustParse(testReceipt),
		PackageID: uuid.MustParse(testPackageID), PackageSlug: "human-resources",
		PackageVersion: "1.0.0", ReleaseID: uuid.MustParse(testReleaseID),
		Component: contract.Component{ID: testComponent},
		State:     contract.ProvisionStaged, CreatedAt: now, UpdatedAt: now,
	}
	receipt := receiptFromSnapshot(snapshot)
	if receipt.Ready || !receipt.CompletedAt.Equal(now) ||
		!receipt.RetryAt.Equal(now.Add(time.Second)) {
		t.Fatalf("pending receipt claimed readiness: %#v", receipt)
	}
	verified := verifiedRequest{
		request:   fixtureRequest([]byte(`{"contract":"fixture.v1"}`), "human-resources"),
		packageID: uuid.MustParse(testPackageID), releaseID: uuid.MustParse(testReleaseID),
		materialization: snapshot.Materialization,
	}
	snapshot.Component = verified.request.Component
	snapshot.RequestSHA256 = sha256.Sum256([]byte("request"))
	snapshot.IdempotencyKey = verified.request.IdempotencyKey
	if !snapshotMatches(snapshot, verified, snapshot.RequestSHA256) {
		t.Fatal("valid pending snapshot was rejected")
	}
}

func fixtureRequest(document []byte, packageSlug string) contract.ProvisionRequest {
	return contract.ProvisionRequest{
		PackageID: testPackageID, PackageSlug: packageSlug,
		PackageVersion: "1.0.0", ReleaseID: testReleaseID,
		Component: contract.Component{
			ID: testComponent, Owner: "workflow", Contract: "fixture.v1",
			ContentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Generation:    85,
		},
		Document: document, IdempotencyKey: "fixture-key",
	}
}

func fixtureSnapshot(
	verified verifiedRequest,
	requestSHA [32]byte,
) Snapshot {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	return Snapshot{
		ReceiptID: uuid.MustParse(testReceipt),
		PackageID: verified.packageID, PackageSlug: verified.request.PackageSlug,
		PackageVersion: verified.request.PackageVersion,
		ReleaseID:      verified.releaseID, Component: verified.request.Component,
		Materialization: verified.materialization.Clone(),
		IdempotencyKey:  verified.request.IdempotencyKey,
		RequestSHA256:   requestSHA, State: contract.ProvisionAvailable,
		ReadyAt: now, CreatedAt: now, UpdatedAt: now,
	}
}

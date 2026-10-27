package iam

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type importClaimViewFixture struct {
	tenantID, jobID, attemptID, leaseOwnerID uuid.UUID
	tenantScoped                             bool
	kind, reference, digest, effectKey       string
	revision, attempt, maximum               int
	effect                                   jobsv1.EffectClass
	leaseGeneration                          int64
	hasEffectKey                             bool
}

func (claim importClaimViewFixture) AttemptID() uuid.UUID { return claim.attemptID }
func (claim importClaimViewFixture) JobID() uuid.UUID     { return claim.jobID }
func (claim importClaimViewFixture) TenantID() (uuid.UUID, bool) {
	return claim.tenantID, claim.tenantScoped
}
func (claim importClaimViewFixture) Kind() string                    { return claim.kind }
func (claim importClaimViewFixture) PayloadRevision() int            { return claim.revision }
func (claim importClaimViewFixture) PayloadReference() string        { return claim.reference }
func (claim importClaimViewFixture) PayloadSHA256() string           { return claim.digest }
func (claim importClaimViewFixture) EffectClass() jobsv1.EffectClass { return claim.effect }
func (claim importClaimViewFixture) EffectIdempotencyKey() (string, bool) {
	return claim.effectKey, claim.hasEffectKey
}
func (claim importClaimViewFixture) Attempt() int            { return claim.attempt }
func (claim importClaimViewFixture) MaximumAttempts() int    { return claim.maximum }
func (claim importClaimViewFixture) LeaseOwnerID() uuid.UUID { return claim.leaseOwnerID }
func (claim importClaimViewFixture) LeaseGeneration() int64  { return claim.leaseGeneration }

func TestSystemTransferImportClaimBindsTenantResourceAndLease(t *testing.T) {
	claim := canonicalImportClaim(t, publicresource.Roles)
	decoded, err := decodeSystemTransferImportClaim(claim, SystemImportJobPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if decoded.TenantID != claim.tenantID || decoded.JobID != claim.jobID ||
		decoded.Resource != publicresource.Roles || decoded.LeaseGeneration != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
	claim.tenantID = uuid.MustParse("019be100-0000-7000-8000-000000000099")
	decoded, err = decodeSystemTransferImportClaim(claim, SystemImportJobPolicy())
	if err != nil || decoded.TenantID != claim.tenantID {
		t.Fatalf("tenant-bound decoded=%+v err=%v", decoded, err)
	}
}

func TestSystemTransferImportClaimRejectsTamperAndExternalEffect(t *testing.T) {
	claim := canonicalImportClaim(t, publicresource.Organizations)
	claim.digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := decodeSystemTransferImportClaim(claim, SystemImportJobPolicy()); err == nil {
		t.Fatal("tampered payload digest was accepted")
	}
	claim = canonicalImportClaim(t, publicresource.Organizations)
	claim.hasEffectKey = true
	if _, err := decodeSystemTransferImportClaim(claim, SystemImportJobPolicy()); err == nil {
		t.Fatal("database import accepted an external effect key")
	}
	claim = canonicalImportClaim(t, publicresource.Organizations)
	claim.tenantScoped = false
	if _, err := decodeSystemTransferImportClaim(claim, SystemImportJobPolicy()); err == nil {
		t.Fatal("global import claim was accepted")
	}
}

func canonicalImportClaim(t *testing.T,
	resource publicresource.Resource,
) importClaimViewFixture {
	t.Helper()
	transferID := uuid.MustParse("019be100-0000-7000-8000-000000000001")
	reference, err := systemtransfer.JobReference(resource, transferID)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := systemtransfer.JobReferenceSHA256(reference)
	if err != nil {
		t.Fatal(err)
	}
	return importClaimViewFixture{
		tenantID:     uuid.MustParse("019be100-0000-7000-8000-000000000002"),
		jobID:        uuid.MustParse("019be100-0000-7000-8000-000000000003"),
		attemptID:    uuid.MustParse("019be100-0000-7000-8000-000000000004"),
		leaseOwnerID: uuid.MustParse("019be100-0000-7000-8000-000000000005"),
		tenantScoped: true, kind: SystemImportJobKind, reference: reference,
		digest: digest, revision: 1, attempt: 1, maximum: 8,
		effect: jobsv1.EffectDatabaseTransactional, leaseGeneration: 1,
	}
}

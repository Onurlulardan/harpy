package iam

import (
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestSystemExportReconcileClaimBindsDedicatedExternalJob(t *testing.T) {
	claim := canonicalExportReconcileClaim(t, publicresource.Users)
	decoded, err := decodeSystemTransferExportReconcileClaim(
		claim, SystemExportReconcileJobPolicy())
	if err != nil || decoded.ExportID.String() != claim.effectKey[len(systemExportReconcileEffectPrefix):] ||
		decoded.Resource != publicresource.Users || decoded.Attempt != 1 ||
		decoded.MaximumAttempts != systemtransfer.ExportReconciliationMaximumAttempts {
		t.Fatalf("decoded = %#v, err = %v", decoded, err)
	}
}

func TestSystemExportReconcileClaimRejectsOriginalExportAndEffectKeyDrift(t *testing.T) {
	claim := canonicalExportReconcileClaim(t, publicresource.Roles)
	claim.kind = SystemExportJobKind
	if _, err := decodeSystemTransferExportReconcileClaim(
		claim, SystemExportReconcileJobPolicy()); err == nil {
		t.Fatal("original export job was accepted for reconciliation")
	}
	claim = canonicalExportReconcileClaim(t, publicresource.Roles)
	claim.effectKey += "-other"
	if _, err := decodeSystemTransferExportReconcileClaim(
		claim, SystemExportReconcileJobPolicy()); err == nil {
		t.Fatal("drifted reconciliation effect key was accepted")
	}
	claim = canonicalExportReconcileClaim(t, publicresource.Roles)
	claim.hasEffectKey = false
	if _, err := decodeSystemTransferExportReconcileClaim(
		claim, SystemExportReconcileJobPolicy()); err == nil {
		t.Fatal("external reconciliation without effect key was accepted")
	}
}

func canonicalExportReconcileClaim(t *testing.T,
	resource publicresource.Resource,
) importClaimViewFixture {
	t.Helper()
	transferID := uuid.MustParse("019be100-0000-7000-8000-000000000011")
	reference, err := systemtransfer.JobReference(resource, transferID)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := systemtransfer.JobReferenceSHA256(reference)
	if err != nil {
		t.Fatal(err)
	}
	return importClaimViewFixture{
		tenantID:     uuid.MustParse("019be100-0000-7000-8000-000000000012"),
		jobID:        uuid.MustParse("019be100-0000-7000-8000-000000000013"),
		attemptID:    uuid.MustParse("019be100-0000-7000-8000-000000000014"),
		leaseOwnerID: uuid.MustParse("019be100-0000-7000-8000-000000000015"),
		tenantScoped: true, kind: SystemExportReconcileJobKind,
		reference: reference, digest: digest, revision: 1, attempt: 1, maximum: 8,
		effect: jobsv1.EffectExternal, leaseGeneration: 1, hasEffectKey: true,
		effectKey: systemExportReconcileEffectPrefix + transferID.String(),
	}
}

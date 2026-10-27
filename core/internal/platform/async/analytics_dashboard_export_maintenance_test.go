package async

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func analyticsMaintenanceCandidate(kind string) AnalyticsDashboardExportMaintenanceUnknownCandidate {
	candidate := AnalyticsDashboardExportMaintenanceUnknownCandidate{
		Kind: kind, TenantID: analyticsExportTestID("000000000001"),
		TenantSchema: "tenant_a", ExportID: analyticsExportTestID("000000000002"),
		AuthorityGroupID: "analytics", AuthorityGroupGeneration: 2,
		CutoverGeneration: 3, JobID: analyticsExportTestID("000000000003"),
		AttemptID: analyticsExportTestID("000000000004"), AttemptNumber: 1,
		LeaseGeneration:      1,
		OriginalLeaseOwnerID: analyticsExportTestID("000000000005"),
		EffectID:             uuid.NewSHA1(uuid.NameSpaceOID, []byte("maintenance-effect")),
		UnknownAt:            time.Unix(1700000000, 123000000).UTC(),
	}
	candidate.EffectRequestSHA256[0] = 9
	return candidate
}

func TestAnalyticsDashboardExportMaintenanceCandidateIsClosed(t *testing.T) {
	deleteCandidate := analyticsMaintenanceCandidate(AnalyticsDashboardExportDeleteJobKind)
	if !deleteCandidate.Valid() {
		t.Fatal("valid delete maintenance candidate was rejected")
	}
	expireCandidate := deleteCandidate
	expireCandidate.Kind = AnalyticsDashboardExportExpireJobKind
	expireCandidate.AttemptNumber = 8
	expireCandidate.LeaseGeneration = 8
	if !expireCandidate.Valid() {
		t.Fatal("valid expiry maintenance candidate was rejected")
	}
	expireCandidate.AttemptNumber = 9
	if expireCandidate.Valid() {
		t.Fatal("expiry attempt above the exact maximum was accepted")
	}
	deleteCandidate.EffectID = analyticsExportTestID("000000000006")
	if deleteCandidate.Valid() {
		t.Fatal("UUIDv7 maintenance effect identity was accepted")
	}
}

func TestAnalyticsDashboardExportMaintenanceCandidateRejectsLeaseDrift(t *testing.T) {
	candidate := analyticsMaintenanceCandidate(AnalyticsDashboardExportDeleteJobKind)
	candidate.LeaseGeneration = candidate.AttemptNumber + 1
	if candidate.Valid() {
		t.Fatal("maintenance lease generation drift was accepted")
	}
	if _, err := AnalyticsDashboardExportMaintenanceReceiptID(candidate); err == nil {
		t.Fatal("receipt identity accepted maintenance lease generation drift")
	}
	_, err := AnalyticsDashboardExportMaintenanceReceiptSHA256(
		candidate, uuid.NewSHA1(uuid.NameSpaceOID, []byte("drifted-receipt")),
		"deleted", candidate.UnknownAt.Add(time.Second),
	)
	if err == nil {
		t.Fatal("receipt digest accepted maintenance lease generation drift")
	}
}

func TestAnalyticsDashboardExportMaintenanceReceiptUsesProductionIdentity(t *testing.T) {
	candidate := analyticsMaintenanceCandidate(AnalyticsDashboardExportDeleteJobKind)
	receiptID, err := AnalyticsDashboardExportMaintenanceReceiptID(candidate)
	want := uuid.NewSHA1(uuid.MustParse("8e7a6d02-df6d-4327-b179-160296b50b8e"),
		[]byte(candidate.EffectID.String()+"\x00ANALYTICS_EXPORT_DELETED"))
	if err != nil || receiptID != want || receiptID.Version() != uuid.Version(5) {
		t.Fatalf("production receipt identity mismatch: id=%s want=%s err=%v",
			receiptID, want, err)
	}
	resolvedAt := candidate.UnknownAt.Add(time.Second)
	digest, err := AnalyticsDashboardExportMaintenanceReceiptSHA256(
		candidate, receiptID, "deleted", resolvedAt,
	)
	if err != nil || digest == [32]byte{} {
		t.Fatalf("canonical receipt digest failed: digest=%x err=%v", digest, err)
	}
	_, err = AnalyticsDashboardExportMaintenanceReceiptSHA256(
		candidate, candidate.EffectID, "deleted", resolvedAt,
	)
	if err == nil {
		t.Fatal("forged receipt identity was accepted")
	}
}

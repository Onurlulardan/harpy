package async

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func analyticsExportTestID(suffix string) uuid.UUID {
	return uuid.MustParse("01900000-0000-7000-8000-" + suffix)
}

func TestAnalyticsDashboardExportProtocolEvidence(t *testing.T) {
	candidate := AnalyticsDashboardExportCandidate{
		TenantID: analyticsExportTestID("000000000001"), TenantSchema: "tenant_a",
		ExportID: analyticsExportTestID("000000000002"),
		JobID:    analyticsExportTestID("000000000003"), AttemptGeneration: 1,
	}
	if !candidate.Valid() {
		t.Fatal("valid dashboard export candidate was rejected")
	}
	candidate.TenantSchema = "Tenant-A"
	if candidate.Valid() {
		t.Fatal("unsafe tenant schema was accepted")
	}
	unknown := AnalyticsDashboardExportUnknownCandidate{
		TenantID: analyticsExportTestID("000000000001"), TenantSchema: "tenant_a",
		ExportID:  analyticsExportTestID("000000000002"),
		JobID:     analyticsExportTestID("000000000003"),
		AttemptID: analyticsExportTestID("000000000004"), AttemptGeneration: 3,
		LeaseOwnerID: analyticsExportTestID("000000000005"),
		EffectID:     analyticsExportTestID("000000000006"),
		UnknownAt:    time.Unix(1, 0).UTC(),
	}
	unknown.EffectRequestSHA256[0] = 1
	if !unknown.Valid() {
		t.Fatal("valid unknown evidence was rejected")
	}
	unknown.AttemptGeneration = 4
	if unknown.Valid() {
		t.Fatal("generation beyond max attempts was accepted")
	}
	unknown.AttemptGeneration = 1
	unknown.EffectRequestSHA256 = [32]byte{}
	if unknown.Valid() {
		t.Fatal("all-zero effect request digest was accepted")
	}
}

func TestAnalyticsDashboardExportSpoolStateIsClosed(t *testing.T) {
	for _, state := range []AnalyticsDashboardExportSpoolState{
		AnalyticsDashboardExportSpoolInactive,
		AnalyticsDashboardExportSpoolRunning,
		AnalyticsDashboardExportSpoolUnknown,
	} {
		if !state.Valid() {
			t.Fatalf("valid spool state %q was rejected", state)
		}
	}
	if AnalyticsDashboardExportSpoolState("queued").Valid() {
		t.Fatal("unknown spool state was accepted")
	}
}

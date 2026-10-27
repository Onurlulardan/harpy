package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestSuccessorFinalizationMigrationsArePairedAndForwardOnly(t *testing.T) {
	for _, version := range []string{"20260729100011", "20260729100012"} {
		source := successorFinalizationMigrationPair(t, version)["fresh"]
		if strings.Count(source, "-- +goose Up") != 1 ||
			strings.Contains(source, "-- +goose Down") ||
			strings.Contains(source, "CREATE OR REPLACE") {
			t.Fatalf("%s fresh migration is not immutable and forward-only", version)
		}
	}
}

func TestSuccessorFinalizationBindsRouteAuthorityAndEvidence(t *testing.T) {
	apply := successorFinalizationMigrationPair(t, "20260729100011")["fresh"]
	for _, marker := range []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"'[\"billing\"]'::jsonb",
		"operation.scope_kind='global' AND operation.group_id='billing'",
		"adoption.state='accepted' AND adoption.item_count=0",
		"zero_receipt.queue_count=0",
		"event.event_kind='request'",
		"INSERT INTO platform.cutover_transition_events",
		"UPDATE platform.domain_cutovers",
		"UPDATE platform.cutover_groups",
		"platform.analytics_finalize_authority_successor_v1(",
		"platform.billing_finalize_authority_successor_v1(",
		"p_operation_id,p_run_id,p_acceptance_digest",
		"platform.workflow_finalize_authority_successor_v1(",
	} {
		if !strings.Contains(apply, marker) {
			t.Fatalf("successor group finalization lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"INSERT INTO platform.analytics_tenant_authorities",
		"INSERT INTO platform.billing_tenant_authorities",
		"INSERT INTO platform.workflow_tenant_authorities",
		"TO harpy_cutover_finalizer",
	} {
		if strings.Contains(apply, forbidden) {
			t.Fatalf("successor group finalization contains %q", forbidden)
		}
	}
}

func TestSuccessorFinalizerRequiresProtectedPairBeforeSwitch(t *testing.T) {
	attach := successorFinalizationMigrationPair(t, "20260729100012")["fresh"]
	for _, marker := range []string{
		"SECURITY INVOKER",
		"cutover_authority.finalization_receipts%ROWTYPE",
		"cutover_authority.verify_finalization_receipt_v1(",
		"cutover_finalize_successor_group_v1(",
		"NEW.operation_id,NEW.run_id,v_receipt.terminal_evidence_sha256",
		"AFTER UPDATE OF state ON platform.cutover_operations",
		"ENABLE ALWAYS TRIGGER",
		"TO harpy_cutover_authority_owner",
	} {
		if !strings.Contains(attach, marker) {
			t.Fatalf("successor finalizer attachment lacks %q", marker)
		}
	}
	if strings.Contains(attach, "TO harpy_cutover_finalizer") {
		t.Fatal("successor finalizer opened a direct finalizer execution surface")
	}
}

func successorFinalizationMigrationPair(
	t *testing.T,
	version string,
) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pair := map[string]string{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if strings.HasPrefix(name, version+"_") && strings.HasSuffix(name, ".fresh.sql") {
			pair["fresh"] = string(document.Content())
		}
	}
	if pair["fresh"] == "" {
		t.Fatalf("fresh migration %s did not load", version)
	}
	return pair
}

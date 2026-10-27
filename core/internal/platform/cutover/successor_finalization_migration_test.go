package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestSuccessorFinalizationMigrationsArePairedAndForwardOnly(t *testing.T) {
	for _, version := range []string{"20260729100011", "20260729100012"} {
		pair := successorFinalizationMigrationPair(t, version)
		if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
			t.Fatalf("%s migration is not byte-identical and paired", version)
		}
		if strings.Count(pair["fresh"], "-- +goose Up") != 1 ||
			strings.Contains(pair["fresh"], "-- +goose Down") ||
			strings.Contains(pair["fresh"], "CREATE OR REPLACE") {
			t.Fatalf("%s migration is not immutable and forward-only", version)
		}
	}
}

func TestSuccessorFinalizationBindsRouteAuthorityAndEvidence(t *testing.T) {
	apply := successorFinalizationMigrationPair(t, "20260729100011")["fresh"]
	for _, marker := range []string{
		"SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"'[\"billing\",\"forms\"]'::jsonb",
		"operation.scope_kind='global' AND operation.group_id='billing'",
		"adoption.state='accepted' AND adoption.item_count=0",
		"zero_receipt.queue_count=0",
		"event.event_kind='request'",
		"INSERT INTO platform.cutover_transition_events",
		"UPDATE platform.domain_cutovers",
		"UPDATE platform.cutover_groups",
		"platform.analytics_finalize_authority_successor_v1(",
		"platform.billing_finalize_authority_successor_v1(",
		"p_operation_id,p_run_id,NULL,NULL,p_acceptance_digest",
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
		"pg_catalog.pg_advisory_xact_lock",
		"FROM platform.cutover_groups AS group_row",
		"group_row.generation=NEW.cutover_generation",
		"billing forms serialization authority differs",
		"IF NOT FOUND THEN RETURN NEW; END IF;",
		"v_companion_receipt.terminal_evidence_sha256",
		"billing forms protected receipts differ",
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
		if !strings.HasPrefix(name, version+"_") {
			continue
		}
		origin := "fresh"
		if strings.Contains(name, ".v1_adopted.sql") {
			origin = "v1_adopted"
		}
		pair[origin] = string(document.Content())
	}
	return pair
}

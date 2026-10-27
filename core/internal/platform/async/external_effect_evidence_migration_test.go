package async

import (
	"bytes"
	"strings"
	"testing"
)

func TestExternalEffectEvidenceMigrationFreezesOwnerBoundary(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	const migration = "20260715022010_create_external_effect_evidence_v1"
	var fresh []byte
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/") &&
			strings.Contains(document.Path(), migration) {
			fresh = document.Content()
		}
	}
	if len(fresh) == 0 {
		t.Fatal("fresh external effect evidence migration did not load")
	}
	if lines := bytes.Count(fresh, []byte{'\n'}) + 1; lines > 200 {
		t.Fatalf("external effect evidence migration has %d lines", lines)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.platform_async_external_effect_evidence_v1(",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"claim_attempt.lease_generation = p_claim_lease_generation",
		"claim_job.payload_revision = p_claim_payload_revision",
		"claim_job.payload_ref = p_claim_payload_reference",
		"claim_job.payload_sha256 = p_claim_payload_sha256",
		"job.dedupe_key = p_expected_dedupe_key",
		"job.effect_idempotency_key = p_expected_effect_key",
		"receipt_row.causation_id <> p_causation_id",
		"p_claim_kind OPERATOR(pg_catalog.~)",
		"pg_catalog.left(claim_job.dedupe_key",
		"MESSAGE = 'external effect terminal job has no attempt'",
		"attempt_row.attempt_id IS NULL",
		"REVOKE ALL ON FUNCTION", "TO harpy_runtime;",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("external effect evidence migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE FUNCTION", "TO harpy_v2_runtime",
		"pg_catalog.coalesce", "pg_catalog.least",
	} {
		if strings.Contains(strings.ToLower(source), strings.ToLower(forbidden)) {
			t.Fatalf("external effect evidence migration contains %q", forbidden)
		}
	}
}

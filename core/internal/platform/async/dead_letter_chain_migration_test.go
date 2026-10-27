package async

import (
	"strings"
	"testing"
)

func TestDeadLetterChainMigrationsFreezeActionAndRedactionGuards(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string][]byte{
		"20260715021130": nil,
		"20260715021140": nil,
	}
	for _, document := range documents {
		version := migrationVersion(t, document.Path())
		if _, wanted := versions[version]; !wanted {
			continue
		}
		if !strings.Contains(document.Path(), "/fresh/") {
			continue
		}
		versions[version] = document.Content()
	}
	for version, fresh := range versions {
		if len(fresh) == 0 {
			t.Fatalf("fresh dead-letter migration %s did not load", version)
		}
	}

	chain := string(versions["20260715021130"])
	chainMarkers := []string{
		"platform_async_validate_dead_letter_chain_v1",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"NEW.writer_origin IS DISTINCT FROM 'v2'",
		"NEW.consumer_name IS NULL AND NEW.consumer_version IS NULL",
		"NEW.consumer_name IS NOT NULL AND NEW.consumer_version IS NOT NULL",
		"FOR UPDATE", "pg_catalog.max(entry.entry_sequence)::bigint + 1",
		"pg_catalog.clock_timestamp() >= root_row.full_payload_expires_at",
		"NEW.authority_group_generation", "root_row.cutover_generation",
		"NEW.source_sha256", "root_row.redacted_error_sha256",
		"NEW.redacted_envelope", "root_row.metadata_expires_at",
		"CREATE TRIGGER dead_letters_chain_insert_guard",
		"ENABLE ALWAYS TRIGGER dead_letters_chain_insert_guard",
		"REVOKE ALL ON FUNCTION",
	}
	for _, marker := range chainMarkers {
		if !strings.Contains(chain, marker) {
			t.Fatalf("dead-letter chain migration lacks %q", marker)
		}
	}

	redaction := string(versions["20260715021140"])
	redactionMarkers := []string{
		"platform_async_guard_dead_letter_redaction_v1",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"OLD.record_kind IS DISTINCT FROM 'root'",
		"expected.payload_canonical := NULL",
		"NEW.payload_redacted_at := redaction_time",
		"expected.payload_redacted_at := redaction_time",
		"NEW IS DISTINCT FROM expected",
		"CREATE TRIGGER dead_letters_redaction_guard",
		"BEFORE UPDATE ON platform.dead_letters",
		"ENABLE ALWAYS TRIGGER dead_letters_redaction_guard",
		"REVOKE ALL ON FUNCTION",
	}
	for _, marker := range redactionMarkers {
		if !strings.Contains(redaction, marker) {
			t.Fatalf("dead-letter redaction migration lacks %q", marker)
		}
	}
}

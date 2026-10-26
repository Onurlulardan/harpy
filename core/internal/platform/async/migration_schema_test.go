package async

import (
	"strings"
	"testing"
)

func TestMigrationSchemaFreezesSliceAContract(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	fresh := map[string]string{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/") {
			fresh[migrationVersion(t, document.Path())] = string(document.Content())
		}
	}
	expected := map[string][]string{
		"20260715020000": {
			"CREATE FUNCTION platform.platform_async_reject_mutation()",
			"SET search_path = pg_catalog", "REVOKE ALL ON FUNCTION",
		},
		"20260715020100": {
			"CREATE TABLE platform.jobs", "authority_group_generation bigint",
			"'database_transactional', 'external'", "'queued', 'running', 'succeeded'",
			"parked_at timestamptz", "JOB_OUTCOME_UNKNOWN",
			"parked_current_group_generation bigint",
			"parked_current_cutover_generation bigint",
			"retain_until timestamptz", "retention_release_revision bigint",
			"CONSTRAINT jobs_retention_check",
			"^[A-Z][A-Z0-9_]{1,63}$",
			"CONSTRAINT jobs_dedupe_unique",
		},
		"20260715020200": {
			"CREATE TABLE platform.job_attempts", "attempt_id uuid PRIMARY KEY",
			"job_attempts_job_attempt_unique", "finished_at timestamptz",
			"effect_receipt_id uuid", "CONSTRAINT job_attempts_terminal_check",
			"error_code IS NULL AND redacted_error_sha256 IS NULL AND\n            effect_receipt_id IS NULL",
			"error_code = 'JOB_OUTCOME_UNKNOWN'",
		},
		"20260715020300": {
			"CREATE TABLE platform.outbox", "actor_id uuid NOT NULL",
			"envelope_contract_version text NOT NULL",
			"business_effect_idempotency_key text NOT NULL",
			"correlation_id text NOT NULL",
			"^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$",
			"payload_canonical bytea",
			"pg_catalog.octet_length(payload_canonical) BETWEEN 2 AND 65536",
			"pg_catalog.sha256(payload_canonical)",
			"pg_catalog.convert_from(payload_canonical, 'UTF8')::jsonb",
			"payload_schema_sha256 text NOT NULL",
			"causation_id uuid NOT NULL", "delivery_state = 'delivered'",
			"event_sequence bigint", "confirm_sha256 text",
		},
		"20260715020400": {
			"CREATE TABLE platform.inbox", "'event_receipt', 'watermark'",
			"consumer_group_id text NOT NULL", "consumer_group_generation bigint NOT NULL",
			"consumer_domain text NOT NULL", "consumer_generation bigint NOT NULL",
			"consumer_writer_origin text NOT NULL CHECK (consumer_writer_origin = 'v2')",
			"CONSTRAINT inbox_consumer_authority_fk",
			"CONSTRAINT inbox_consumer_authority_group_fk",
			"CONSTRAINT inbox_consumer_authority_member_fk",
			"CONSTRAINT inbox_event_unique", "consumer_version, event_id",
			"CONSTRAINT inbox_event_sequence_unique", "causation_id uuid",
			"through_sequence bigint",
		},
		"20260715020500": {
			"CREATE TABLE platform.idempotency_receipts", "'http', 'effect'",
			"pg_catalog.octet_length(response_body) <= 65536",
			"pg_catalog.octet_length(response_headers::text) <= 16384",
			"async_job_id uuid", "idempotency_receipts_async_job_fk",
			"effect_id uuid", "effect_kind text", "effect_revision bigint",
			"provider_reference text",
			"effect_status IN ('succeeded', 'failed')",
			"effect_code ~ '^[A-Z][A-Z0-9_]{1,63}$'",
			"pg_catalog.octet_length(effect_result_ref) BETWEEN 1 AND 2048",
			"job_attempts_effect_receipt_fk",
		},
		"20260715020600": {
			"CREATE TABLE platform.dead_letters",
			"consumer_name text", "consumer_version integer",
			"'root', 'retry_action', 'repair_action'",
			"payload_canonical bytea CHECK (",
			"pg_catalog.octet_length(payload_canonical) BETWEEN 2 AND 65536",
			"redacted_envelope jsonb NOT NULL",
			"pg_catalog.octet_length(redacted_envelope::text) <= 16384",
			"CONSTRAINT dead_letters_payload_hash_check",
			"CONSTRAINT dead_letters_chain_unique", "payload_redacted_at timestamptz",
			"(action_reason IS NULL AND action_ticket IS NULL)",
		},
		"20260715020700": {
			"CREATE FUNCTION platform.platform_async_assert_current_authority_v1(",
			"FROM platform.cutover_groups", "FOR SHARE",
			"FROM platform.domain_cutovers",
			"FROM platform.cutover_group_members",
			"REVOKE ALL ON FUNCTION",
		},
		"20260715020730": {
			"CREATE FUNCTION platform.platform_async_validate_owner_insert_v1()",
			"SECURITY DEFINER", "SESSION_USER", "harpy_adopter",
			"NEW.writer_origin", "platform_async_assert_current_authority_v1(",
			"BEFORE INSERT ON platform.jobs", "ENABLE ALWAYS TRIGGER",
		},
		"20260715020740": {
			"CREATE FUNCTION platform.platform_async_validate_inbox_insert_v1()",
			"NEW.consumer_writer_origin IS DISTINCT FROM 'v2'",
			"FROM platform.outbox AS event_row", "JOIN platform.inbox AS receipt",
			"NEW.producer_domain", "NEW.consumer_domain",
			"NEW.watermark_kind OPERATOR(pg_catalog.=) 'bootstrap'",
			"BEFORE INSERT ON platform.inbox", "ENABLE ALWAYS TRIGGER",
		},
		"20260715020800": {
			"CREATE FUNCTION platform.platform_async_guard_append_only_v1()",
			"CURRENT_USER OPERATOR(pg_catalog.=) 'harpy_owner'",
			"CREATE TRIGGER jobs_delete_guard",
			"CREATE TRIGGER job_attempts_delete_guard",
			"CREATE TRIGGER outbox_delete_guard",
			"CREATE TRIGGER inbox_append_only_guard",
			"CREATE TRIGGER idempotency_receipts_append_only_guard",
			"CREATE TRIGGER dead_letters_delete_guard",
		},
		"20260715021610": {
			"ADD COLUMN parked_at timestamptz",
			"parked_code = 'OUTBOX_AUTHORITY_STALE'",
			"CONSTRAINT outbox_parked_check",
			"CREATE INDEX outbox_stale_authority_idx",
		},
		"20260715021620": {
			"CREATE FUNCTION platform.platform_async_guard_outbox_park_v1()",
			"SESSION_USER", "CURRENT_USER",
			"CREATE TRIGGER outbox_park_guard",
			"ENABLE ALWAYS TRIGGER outbox_park_guard",
		},
		"20260715021630": {
			"CREATE FUNCTION platform.platform_async_park_stale_outbox_v1(",
			"SECURITY DEFINER", "FOR UPDATE",
			"INSERT INTO platform.dead_letters",
			"'stale_generation', 'OUTBOX_AUTHORITY_STALE'",
			"GRANT EXECUTE ON FUNCTION",
		},
	}
	if len(fresh) < len(expected) {
		t.Fatalf("fresh migration count = %d, want at least %d", len(fresh), len(expected))
	}
	for version, markers := range expected {
		source, exists := fresh[version]
		if !exists {
			t.Fatalf("migration %s is missing", version)
		}
		for _, marker := range markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("migration %s lacks %q", version, marker)
			}
		}
		if strings.Contains(source, "-- +goose Down") {
			t.Fatalf("migration %s contains a destructive down section", version)
		}
	}
	for _, version := range []string{"20260715020730", "20260715020740"} {
		if strings.Contains(fresh[version], "BEFORE UPDATE") ||
			strings.Contains(fresh[version], "BEFORE DELETE") {
			t.Fatalf("migration %s insert guard handles another operation", version)
		}
	}
	if strings.Contains(fresh["20260715020200"], "provider_reference") ||
		strings.Contains(fresh["20260715020200"], "outcome_unknown") {
		t.Fatal("job attempt duplicated effect receipt or invented an outcome")
	}
	for _, version := range []string{"20260715020300", "20260715020600"} {
		if strings.Contains(fresh[version], "payload jsonb") {
			t.Fatalf("migration %s stores authoritative payload as jsonb", version)
		}
	}
	for _, version := range []string{
		"20260715020100", "20260715020300", "20260715020400",
		"20260715020500", "20260715020600",
	} {
		if strings.Count(fresh[version], "^[a-z][a-z0-9_]{0,62}$") < 2 {
			t.Fatalf("migration %s lacks bounded authority identifiers", version)
		}
	}
}

func migrationVersion(t *testing.T, path string) string {
	t.Helper()
	base := path[strings.LastIndex(path, "/")+1:]
	if len(base) < 14 {
		t.Fatalf("migration path has no version: %s", path)
	}
	return base[:14]
}

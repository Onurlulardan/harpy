package async

import (
	"strings"
	"testing"
)

func TestTerminalJobDeadLetterMigrationsFreezeAtomicEvidence(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string][]byte{
		"20260715021900": nil,
		"20260715021910": nil,
		"20260715021920": nil,
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
			t.Fatalf("fresh terminal DLQ migration %s did not load", version)
		}
	}

	index := string(versions["20260715021900"])
	for _, marker := range []string{
		"CREATE UNIQUE INDEX dead_letters_terminal_job_root_unique",
		"WHERE record_kind = 'root' AND source_kind = 'job'",
		"reason_class IN ('permanent','delivery_limit','outcome_unknown')",
	} {
		if !strings.Contains(index, marker) {
			t.Fatalf("terminal DLQ unique migration lacks %q", marker)
		}
	}
	if strings.Contains(index, "'stale_generation'") {
		t.Fatal("terminal DLQ unique migration captured stale-generation evidence")
	}

	recorder := string(versions["20260715021910"])
	for _, marker := range []string{
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"v_attempt.outcome <> 'dead'", "FOR SHARE",
		"WHEN NEW.last_error_class = 'permanent' THEN 'permanent'",
		"ELSE 'delivery_limit'", "THEN 'outcome_unknown'",
		"ON CONFLICT DO NOTHING", "GET DIAGNOSTICS v_inserted = ROW_COUNT",
		"root.reason_class IN ('permanent','delivery_limit','outcome_unknown')",
		"pg_catalog.substr(v_root_digest, 9, 4) || '-8'",
		"pg_catalog.substr('89ab'", "pg_catalog.get_byte(",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(recorder, marker) {
			t.Fatalf("terminal DLQ recorder migration lacks %q", marker)
		}
	}
	assertTerminalPayloadJCSOrder(t, recorder)

	binding := string(versions["20260715021920"])
	for _, marker := range []string{
		"AFTER UPDATE OF state ON platform.jobs",
		"OLD.state IS DISTINCT FROM 'dead' AND NEW.state = 'dead'",
		"ENABLE ALWAYS TRIGGER jobs_terminal_dead_letter_guard",
	} {
		if !strings.Contains(binding, marker) {
			t.Fatalf("terminal DLQ binding migration lacks %q", marker)
		}
	}
}

func assertTerminalPayloadJCSOrder(t *testing.T, source string) {
	t.Helper()
	keys := []string{
		`'{"attempt":'`, `,"contract_version":`, `,"job_id":`, `,"kind":`,
		`,"payload_reference":`, `,"payload_revision":`, `,"payload_sha256":`,
		`,"reason_code":`,
	}
	position := -1
	for _, key := range keys {
		next := strings.Index(source[position+1:], key)
		if next < 0 {
			t.Fatalf("terminal DLQ JCS payload lacks ordered key %q", key)
		}
		position += next + 1
	}
}

package async

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalJobDeadLetterMigrationsFreezeAtomicEvidence(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	versions := map[string]map[string][]byte{
		"20260715021900": {},
		"20260715021910": {},
		"20260715021920": {},
	}
	for _, document := range documents {
		version := migrationVersion(t, document.Path())
		if _, wanted := versions[version]; !wanted {
			continue
		}
		origin := "fresh"
		if strings.Contains(document.Path(), "/v1_adopted/") {
			origin = "v1_adopted"
		}
		versions[version][origin] = document.Content()
	}
	for version, origins := range versions {
		fresh, freshOK := origins["fresh"]
		adopted, adoptedOK := origins["v1_adopted"]
		if !freshOK || !adoptedOK || !bytes.Equal(fresh, adopted) {
			t.Fatalf("terminal DLQ migration %s origin pair differs", version)
		}
	}

	index := string(versions["20260715021900"]["fresh"])
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

	recorder := string(versions["20260715021910"]["fresh"])
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

	binding := string(versions["20260715021920"]["fresh"])
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

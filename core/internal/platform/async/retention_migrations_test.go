package async

import (
	"bytes"
	"os"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestRetentionMigrationsFreezePairedPolicies(t *testing.T) {
	files := []string{
		"20260715021300_redact_async_job_payloads_v1",
		"20260715021310_redact_async_outbox_payloads_v1",
		"20260715021320_redact_async_dead_letter_payloads_v1",
		"20260715021400_purge_async_inbox_v1",
		"20260715021410_purge_async_idempotency_receipts_v1",
		"20260715021430_create_async_job_purge_v1",
		"20260715021450_create_async_dead_letter_purge_v1",
	}
	sources := make(map[string]string, len(files))
	for _, name := range files {
		fresh := readRetentionMigration(t, "fresh", name)
		adopted := readRetentionMigration(t, "v1_adopted", name)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("migration %s origin pair differs", name)
		}
		if bytes.Count(fresh, []byte("\n")) > 200 {
			t.Fatalf("migration %s exceeds 200 lines", name)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("migration %s is invalid: %v", name, err)
		}
		source := string(fresh)
		for _, marker := range []string{
			"SECURITY DEFINER", "SET search_path = pg_catalog",
			"pg_catalog.clock_timestamp()", "NOT BETWEEN 1 AND 500",
			"FOR UPDATE", "SKIP LOCKED", "REVOKE ALL ON FUNCTION",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("migration %s lacks %q", name, marker)
			}
		}
		sources[name[:14]] = source
	}

	assertRetentionMarkers(t, sources["20260715021300"], []string{
		"platform_async_redact_job_payloads_v1",
		"job.completed_at <= v_now - interval '30 days'",
	})
	assertRetentionMarkers(t, sources["20260715021310"], []string{
		"platform_async_redact_outbox_payloads_v1",
		"event.delivered_at <= v_now - interval '30 days'",
	})
	assertRetentionMarkers(t, sources["20260715021320"], []string{
		"platform_async_redact_dead_letter_payloads_v1",
		"dead_letter.full_payload_expires_at <= v_now",
	})
	assertRetentionMarkers(t, sources["20260715021400"], []string{
		"platform_async_purge_inbox_v1",
		"inbox_row.recorded_at <= v_now - interval '180 days'",
		"newer.through_sequence > inbox_row.through_sequence",
	})
	assertRetentionMarkers(t, sources["20260715021410"], []string{
		"platform_async_purge_idempotency_receipts_v1",
		"receipt.retain_until <= v_now",
		"receipt.retention_class <> 'async_handle'",
		"job.state IN ('succeeded', 'dead', 'cancelled')",
		"job.completed_at <= v_now - interval '24 hours'",
		"attempt.effect_receipt_id = receipt.receipt_id",
	})
	assertRetentionMarkers(t, sources["20260715021430"], []string{
		"job.state IN ('succeeded', 'dead', 'cancelled')",
		"job.payload_redacted_at IS NOT NULL", "job.retain_until <= v_now",
		"job.retention_release_revision > 0",
		"job.retention_release_sha256 IS NOT NULL",
		"job.retention_released_at IS NOT NULL",
		"attempt.finished_at IS NULL", "receipt.async_job_id = job.id",
		"retry.retry_of_job_id = job.id",
	})
	assertDeleteOrder(t, sources["20260715021430"],
		"DELETE FROM platform.job_attempts", "DELETE FROM platform.jobs")

	assertRetentionMarkers(t, sources["20260715021450"], []string{
		"platform_async_purge_dead_letters_v1",
		"dead_letter.metadata_expires_at <= v_now",
		"dead_letter.payload_redacted_at IS NOT NULL",
		"ORDER BY action.entry_sequence DESC",
		"root.root_id = root.id",
	})
	assertDeleteOrder(t, sources["20260715021450"],
		"DELETE FROM platform.dead_letters AS action",
		"DELETE FROM platform.dead_letters AS root")
}

func readRetentionMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	path := "migrations/" + origin + "/" + name + "." + origin + ".sql"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func assertRetentionMarkers(t *testing.T, source string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("retention migration lacks %q", marker)
		}
	}
}

func assertDeleteOrder(t *testing.T, source, child, parent string) {
	t.Helper()
	childIndex, parentIndex := strings.Index(source, child), strings.Index(source, parent)
	if childIndex < 0 || parentIndex < 0 || childIndex >= parentIndex {
		t.Fatalf("retention deletion order is invalid: %q before %q", child, parent)
	}
}

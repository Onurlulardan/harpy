package billing

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderCommandMigrationsAreExactBoundedPairs(t *testing.T) {
	names := []string{
		"20260717010250_complete_billing_provider_command_ledger_v1",
		"20260717010251_create_billing_provider_command_v1",
		"20260717010253_create_billing_provider_finalizer_v1",
		"20260717010254_guard_billing_provider_command_v1",
	}
	var combined strings.Builder
	for _, name := range names {
		fresh := readProviderMigration(t, "fresh", name+".fresh.sql")
		adopted := readProviderMigration(
			t, "v1_adopted", name+".v1_adopted.sql",
		)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("migration pair %s differs", name)
		}
		if lines := bytes.Count(fresh, []byte("\n")) + 1; lines > 200 {
			t.Fatalf("migration %s has %d lines", name, lines)
		}
		combined.Write(fresh)
	}
	document := combined.String()
	required := []string{
		"ADD COLUMN actor_id uuid",
		"ADD COLUMN active_job_id uuid",
		"ADD COLUMN finalizer_job_id uuid",
		"billing_create_provider_command_v1",
		"platform_async_external_effect_evidence_v1",
		"billing_finalize_provider_command_v1",
		"FOR UPDATE OF command_row",
		"stored.finalizer_job_id IS DISTINCT FROM p_finalizer_job_id",
		"SET finalizer_job_id = p_next_finalizer_job_id",
		"billing_provider_commands_transition_guard",
		"GRANT EXECUTE ON FUNCTION",
	}
	for _, fragment := range required {
		if !strings.Contains(document, fragment) {
			t.Fatalf("provider migration missing %q", fragment)
		}
	}
	for _, forbidden := range []string{
		"pg_catalog.coalesce", "pg_catalog.least", "DROP ", "CASCADE",
	} {
		if strings.Contains(document, forbidden) {
			t.Fatalf("provider migration contains %q", forbidden)
		}
	}
}

func readProviderMigration(
	t *testing.T,
	stream string,
	name string,
) []byte {
	t.Helper()
	path := filepath.Join(
		"migrations", "platform", stream, name,
	)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

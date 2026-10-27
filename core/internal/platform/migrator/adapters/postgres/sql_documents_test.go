package postgres

import (
	"strings"
	"testing"
)

func TestEmbeddedSQLCarriesExactImmutableHistoryContract(t *testing.T) {
	documents := []string{
		tableExistsSQL, createSchemaSQL, createHistoryTableSQL,
		createRejectMutationFunctionSQL, createRejectMutationTriggerSQL,
		insertAnchorSQL, insertMigrationSQL, getMigrationSQL,
		getLatestVersionSQL, listMigrationsSQL,
	}
	for index, document := range documents {
		if strings.TrimSpace(document) == "" {
			t.Fatalf("embedded SQL document %d is empty", index)
		}
	}
	for _, member := range []string{
		"version", "owner", "scope", "origin", "selected_path", "selected_sha256",
		"run_id", "target_database_generation", "next_minimum_runtime_generation",
		"applied_at",
	} {
		if !strings.Contains(createHistoryTableSQL, member) {
			t.Fatalf("history table omitted %s", member)
		}
	}
	if !strings.Contains(createRejectMutationTriggerSQL, "BEFORE UPDATE OR DELETE") {
		t.Fatal("history mutation trigger is not fail-closed")
	}
}

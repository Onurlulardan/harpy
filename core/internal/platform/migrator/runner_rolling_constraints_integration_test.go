//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertReleaseGenerationConstraints(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	expected := map[string][]string{
		"release_generations_database_floor": {
			"minimum_database_generation", "database_generation", "<=",
		},
		"release_generations_runtime_floor": {
			"minimum_runtime_generation", "runtime_generation", "<=",
		},
		"release_generations_database_manifest_fk": {
			"FOREIGN KEY", "database_generation", "database_generation_manifests",
		},
	}
	for name, fragments := range expected {
		var definition string
		err := database.QueryRowContext(ctx, `SELECT pg_catalog.pg_get_constraintdef(oid)
FROM pg_catalog.pg_constraint
WHERE conrelid='platform.release_generations'::pg_catalog.regclass
  AND conname=$1`, name).Scan(&definition)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(definition, fragment) {
				t.Fatalf("%s definition %q lacks %q", name, definition, fragment)
			}
		}
	}
	var deferrable, initiallyDeferred bool
	err := database.QueryRowContext(ctx, `SELECT tgdeferrable,tginitdeferred
FROM pg_catalog.pg_trigger
WHERE tgrelid='platform.release_generations'::pg_catalog.regclass
  AND tgname='release_generations_require_active_contract'`).Scan(
		&deferrable, &initiallyDeferred,
	)
	if err != nil || !deferrable || !initiallyDeferred {
		t.Fatalf("release finalization trigger is not deferred: %t %t %v",
			deferrable, initiallyDeferred, err)
	}
}

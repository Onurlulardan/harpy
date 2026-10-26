//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMFoundationHasNoRuntimeDML(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, role := range []string{"harpy_runtime", "harpy_adopter"} {
		var granted int
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE relation.relkind IN ('r', 'p')
  AND relation.relname LIKE 'iam\_%' ESCAPE '\'
  AND namespace.nspname IN ('platform', 't_iam_a', 't_iam_b')
  AND (
      pg_catalog.has_table_privilege($1, relation.oid, 'SELECT') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'INSERT') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'UPDATE') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'DELETE') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'TRUNCATE') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'REFERENCES') OR
      pg_catalog.has_table_privilege($1, relation.oid, 'TRIGGER')
  )`, role).Scan(&granted); err != nil {
			t.Fatalf("read %s IAM table grants: %v", role, err)
		}
		if granted != 0 {
			t.Fatalf("%s has direct privilege on %d IAM tables", role, granted)
		}
	}
}

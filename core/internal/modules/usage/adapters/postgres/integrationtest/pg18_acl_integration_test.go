//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestUsageACLOnFreshAndAdoptedPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startUsageDatabase(t, ctx)

	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetAndApplyUsage(t, ctx, database, origin)

			var directTableAccess bool
			err := database.QueryRowContext(ctx, `
SELECT COALESCE(pg_catalog.bool_or(pg_catalog.has_table_privilege(
           roles.role_name, relations.oid, privileges.privilege)), false)
FROM pg_catalog.pg_class AS relations
JOIN pg_catalog.pg_namespace AS namespaces
  ON namespaces.oid = relations.relnamespace
CROSS JOIN (VALUES ('harpy_runtime'), ('harpy_adopter')) AS roles(role_name)
CROSS JOIN (VALUES ('SELECT'), ('INSERT'), ('UPDATE'), ('DELETE'),
                   ('TRUNCATE'), ('REFERENCES'), ('TRIGGER')) AS privileges(privilege)
WHERE namespaces.nspname = 'platform'
  AND relations.relname IN (
      'usage_request_facts', 'usage_legal_holds', 'usage_retention_receipts',
      'usage_adoption_runs', 'usage_adoption_mappings',
      'usage_adoption_page_receipts')`).Scan(&directTableAccess)
			if err != nil {
				t.Fatal(err)
			}
			if directTableAccess {
				t.Fatal("runtime or adopter received direct usage table access")
			}

			var runtimeExecute, adopterExecute, publicExecute int
			err = database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
           'harpy_runtime', routines.oid, 'EXECUTE')),
       pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
           'harpy_adopter', routines.oid, 'EXECUTE')),
       pg_catalog.count(*) FILTER (WHERE pg_catalog.has_function_privilege(
           0::oid, routines.oid, 'EXECUTE'))
FROM pg_catalog.pg_proc AS routines
JOIN pg_catalog.pg_namespace AS namespaces
  ON namespaces.oid = routines.pronamespace
WHERE namespaces.nspname = 'platform'
  AND routines.proname LIKE 'usage\_%\_v1' ESCAPE '\'`).Scan(
				&runtimeExecute, &adopterExecute, &publicExecute)
			if err != nil {
				t.Fatal(err)
			}
			if runtimeExecute != 9 || adopterExecute != 9 || publicExecute != 0 {
				t.Fatalf("usage routine ACL drift: runtime=%d adopter=%d public=%d",
					runtimeExecute, adopterExecute, publicExecute)
			}
		})
	}
}

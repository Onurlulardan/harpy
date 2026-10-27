//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsCatalogRuntimePrivileges(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var allowed bool
	err := database.QueryRowContext(ctx, `SELECT
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_project_v1(name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_project_v2(name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_component_apply_v1(uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_component_apply_v2(uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_component_blockers_v1(uuid,uuid,uuid,bigint,bigint,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_publication_create_v1(name,uuid,uuid,text,bytea,uuid,text,uuid,text,text,bytea,timestamptz)',
'EXECUTE') AND
has_function_privilege('harpy_runtime',
't_analytics_a.analytics_reference_register_v1(name,uuid,uuid,text,uuid,text,text,uuid,uuid,text,timestamptz)',
'EXECUTE') AND
has_table_privilege('harpy_runtime',
't_analytics_a.analytics_catalog_component_candidates','SELECT')`,
	).Scan(&allowed)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("harpy_runtime lacks an Analytics Catalog runtime privilege")
	}
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func assertIAMPermissionRequestCatalogRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	resources := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT COALESCE(pg_catalog.string_agg(resource_key,',' ORDER BY resource_key),'')
FROM iam_permission_request_resource_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'all','',NULL,101,NULL
)`, iamPermissionRequester)
	if resources != "crm.contacts,form.orders" {
		t.Fatalf("requestable catalog resources = %q", resources)
	}
	actions := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT COALESCE(pg_catalog.string_agg(action_key,',' ORDER BY action_key),'')
FROM iam_permission_request_action_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'',NULL,101
)`, iamPermissionRequester)
	if actions != "read" {
		t.Fatalf("requestable catalog actions = %q", actions)
	}
	missing := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT COALESCE(pg_catalog.string_agg(resource_key,',' ORDER BY resource_key),'')
FROM iam_permission_request_resource_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'all','',NULL,101,'read'
)`, iamPermissionRequester)
	if missing != "crm.contacts" {
		t.Fatalf("missing permission resources = %q", missing)
	}
	assertIAMPermissionRequestCatalogDenials(t, ctx, database, schema)
}

func assertIAMPermissionRequestCatalogDenials(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	other := "t_iam_a"
	if schema == other {
		other = "t_iam_b"
	}
	err := rejectIAMPGFirst(t, ctx, database, schema, `
SELECT * FROM iam_permission_request_resource_catalog_v1(
    $1::name,$2,'all','',NULL,10,NULL
)`, other, iamPermissionRequester)
	requireIAMCatalogSQLState(t, err, "42501")
	err = rejectIAMPGFirst(t, ctx, database, schema, `
SELECT * FROM iam_permission_request_action_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'',NULL,10
)`, "ffffffff-ffff-4fff-8fff-ffffffffffff")
	requireIAMCatalogSQLState(t, err, "42501")
	err = rejectIAMPGFirst(t, ctx, database, schema, `
SELECT * FROM iam_permission_request_resource_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'form','',NULL,10,'unknown'
)`, iamPermissionRequester)
	requireIAMCatalogSQLState(t, err, "P0002")
}

func requireIAMCatalogSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("catalog denial = %s, want SQLSTATE %s", iamPGErrorDetail(err), want)
	}
}

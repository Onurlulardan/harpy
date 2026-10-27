//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const isolatedLayoutRevision = "dddddddd-dddd-7ddd-8ddd-dddddddddddd"

func assertLayoutAuthorityAndIsolation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	auto := adapterDefinition
	text := strings.Replace(auto, `"widget":"auto"`, `"widget":"text"`, 1)
	createLayoutRevision(t, ctx, database, "layout_a", isolatedLayoutRevision,
		layoutapi.ResourceOrganization, auto)
	createLayoutRevision(t, ctx, database, "layout_b", isolatedLayoutRevision,
		layoutapi.ResourceOrganization, text)
	left := readLayoutRevision(t, ctx, database, "layout_a", getSQL,
		isolatedLayoutRevision)
	right := readLayoutRevision(t, ctx, database, "layout_b", getSQL,
		isolatedLayoutRevision)
	if left.DefinitionSHA256 == right.DefinitionSHA256 ||
		left.Definition.Surfaces[0].Sections[0].Fields[0].Widget != layoutapi.WidgetAuto ||
		right.Definition.Surfaces[0].Sections[0].Fields[0].Widget != layoutapi.WidgetText {
		t.Fatalf("tenant layouts crossed: %+v / %+v", left, right)
	}
	publishLayoutRevision(
		t, ctx, database, "layout_a", isolatedLayoutRevision, 1, auto, layoutNow(11),
	)
	publishLayoutRevision(
		t, ctx, database, "layout_b", isolatedLayoutRevision, 1, text, layoutNow(11),
	)
	leftPublished := readPublishedLayout(
		t, ctx, database, "layout_a", resolveSQL, "organization",
	)
	rightPublished := readPublishedLayout(
		t, ctx, database, "layout_b", resolveSQL, "organization",
	)
	if leftPublished.Revision.DefinitionSHA256 ==
		rightPublished.Revision.DefinitionSHA256 ||
		leftPublished.Revision.ID != isolatedLayoutRevision ||
		rightPublished.Revision.ID != isolatedLayoutRevision {
		t.Fatalf("tenant publications crossed: %+v / %+v", leftPublished, rightPublished)
	}
	_, err := runLayoutRuntimeText(ctx, database, "layout_a", getSQL,
		isolatedLayoutRevision)
	if err != nil {
		t.Fatal(err)
	}
	_, err = runLayoutRuntimeText(ctx, database, "layout_a",
		"SELECT system_resource_layout_get_v1($1::name, $2::uuid)",
		"layout_b", isolatedLayoutRevision)
	assertLayoutSQLState(t, err, "42501")
	assertLayoutRuntimeACL(t, ctx, database)
	assertPresentationRoutineACL(t, ctx, database)
	assertLayoutWriteRoutineACL(t, ctx, database)
	assertLayoutDirectDMLDenied(t, ctx, database)
}

func assertPresentationRoutineACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var publishV5, resolveV2, publishV4, createV3, createV2 bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_publish_v5(name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_publish_v4(name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime','layout_a.system_resource_layout_resolve_v2(name,text)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime','layout_a.system_resource_layout_create_v3(name,uuid,text,text,bytea,timestamptz)',
    'EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime','layout_a.system_resource_layout_create_v2(name,uuid,text,text,bytea,timestamptz)',
    'EXECUTE'
)`).Scan(&publishV5, &publishV4, &resolveV2, &createV3, &createV2)
	if err != nil {
		t.Fatal(err)
	}
	if !publishV5 || publishV4 || !resolveV2 || !createV3 || createV2 {
		t.Fatalf("layout ACL publish5/publish4/resolve2/create3/create2 = %t/%t/%t/%t/%t",
			publishV5, publishV4, resolveV2, createV3, createV2)
	}
}

func assertLayoutRuntimeACL(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var directTable, publicRoutine, internalRoutine, adopterRoutine bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_table_privilege(
    'harpy_runtime','layout_a.system_resource_layout_revisions',
    'SELECT,INSERT,UPDATE,DELETE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_get_v1(name,uuid)','EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_runtime',
    'layout_a.system_resource_layout_revision_json_v1(uuid)','EXECUTE'
),
pg_catalog.has_function_privilege(
    'harpy_adopter',
    'layout_a.system_resource_layout_get_v1(name,uuid)','EXECUTE'
)`).Scan(&directTable, &publicRoutine, &internalRoutine, &adopterRoutine)
	if err != nil {
		t.Fatal(err)
	}
	if directTable || !publicRoutine || internalRoutine || adopterRoutine {
		t.Fatalf("layout ACL table/public/internal/adopter = %t/%t/%t/%t",
			directTable, publicRoutine, internalRoutine, adopterRoutine)
	}
}

func assertLayoutDirectDMLDenied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	connection, transaction, err := beginLayoutRuntimeTx(ctx, database, "layout_a")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var count int
	err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.count(*) FROM system_resource_layout_revisions").Scan(&count)
	_ = transaction.Rollback()
	assertLayoutSQLState(t, err, "42501")
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const legacyLayoutRevisionID = "90909090-9090-4090-8090-909090909090"

func seedLegacyLayoutRevision(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	statement := `SELECT system_resource_layout_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::text,$3::text,$4::bytea,$5::timestamptz
)`
	_, err := runLayoutRuntimeText(
		ctx, database, "layout_a", statement,
		legacyLayoutRevisionID, string(layoutapi.ResourceRole),
		adapterLegacyDefinition, layoutDigest(adapterLegacyDefinition), layoutNow(0),
	)
	if err != nil {
		t.Fatalf("seed legacy layout revision: %v", err)
	}
}

func assertLegacyLayoutCompatibility(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	revision := readLayoutRevision(
		t, ctx, database, "layout_a", getSQL, legacyLayoutRevisionID,
	)
	if revision.Definition.SourceContract != layoutapi.DefinitionContractV1 ||
		len(revision.Definition.Surfaces) != 4 {
		t.Fatalf("legacy layout read = %+v", revision.Definition)
	}
	_, err := runLayoutRuntimeText(
		ctx, database, "layout_a", updateSQL, legacyLayoutRevisionID, int64(1),
		adapterDefinition, layoutDigest(adapterDefinition), layoutNow(1),
	)
	assertLayoutSQLState(t, err, "40001")
	assertLegacyWriteRoutineRevoked(t, ctx, database)
	arguments := layoutPublishArguments(
		legacyLayoutRevisionID, 1, adapterLegacyDefinition, layoutNow(2),
	)
	arguments[3] = "system-resource-presentation-compiler.v1"
	published := readPublishedLayout(
		t, ctx, database, "layout_a", publishSQL, arguments...,
	)
	if published.CompilerVersion != "system-resource-presentation-compiler.v1" ||
		published.Revision.Definition.SourceContract != layoutapi.DefinitionContractV1 {
		t.Fatalf("legacy layout publication = %+v", published)
	}
}

func assertLegacyWriteRoutineRevoked(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	statement := `SELECT system_resource_layout_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::uuid,$2::text,$3::text,$4::bytea,$5::timestamptz
)`
	_, err := runLayoutRuntimeText(
		ctx, database, "layout_a", statement,
		"91919191-9191-4191-8191-919191919191", string(layoutapi.ResourceRole),
		adapterLegacyDefinition, layoutDigest(adapterLegacyDefinition), layoutNow(1),
	)
	assertLayoutSQLState(t, err, "42501")
}

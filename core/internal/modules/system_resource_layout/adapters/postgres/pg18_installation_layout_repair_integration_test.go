//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestInstallationLayoutRepairAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	database := startLayoutPG18(t, ctx, "repair_fresh")
	prepareLayoutPG18(t, ctx, database)
	_, err := database.ExecContext(ctx, `
CREATE SCHEMA layout_missing;
CREATE SCHEMA layout_semantic;
CREATE SCHEMA layout_customer;
CREATE SCHEMA layout_parity_corrupt;
GRANT USAGE ON SCHEMA layout_missing,layout_semantic,layout_customer,
layout_parity_corrupt
TO harpy_runtime;
REVOKE CREATE ON SCHEMA layout_missing,layout_semantic,layout_customer,
layout_parity_corrupt
FROM PUBLIC,harpy_runtime;`)
	if err != nil {
		t.Fatal(err)
	}
	schemas := []string{
		layoutContractSchema, "layout_a", "layout_b",
		"layout_missing", "layout_semantic", "layout_parity_corrupt",
	}
	predecessor, repair := layoutMigrationRange(t,
		"20260730120100_create_layout_repair_codec_v1",
		"20260730120104_repair_installation_layout_publications_v1",
	)
	_, parity := layoutMigrationRange(t,
		"20260730120104_repair_installation_layout_publications_v1",
		"20260802232102_repair_layout_curated_parity_v2",
	)
	parity = parity[1:]
	for _, schema := range schemas {
		applyLayoutMigrationSetPG18(
			t, ctx, database, schema, predecessor,
		)
		seedInstallationLayoutRepairPG18(t, ctx, database, schema)
	}
	seedCustomLayoutRepairPG18(t, ctx, database, "layout_a")
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_customer", predecessor,
	)
	seedCustomLayoutRepairPG18(t, ctx, database, "layout_customer")
	customerSnapshot := layoutRepairSnapshotPG18(
		t, ctx, database, "layout_customer", "publications",
	)
	corruptLayoutRepairSourceSHA(t, ctx, database)
	removeLayoutRepairPublication(t, ctx, database)
	corruptLayoutRepairSemanticCatalog(t, ctx, database)
	oldRows := layoutRepairSnapshotPG18(
		t, ctx, database, layoutContractSchema, "old_revisions",
	)
	oldCatalogs := layoutRepairSnapshotPG18(
		t, ctx, database, layoutContractSchema, "publications",
	)
	applyLayoutMigrationSetPG18(
		t, ctx, database, layoutContractSchema, repair,
	)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_a", repair,
	)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_parity_corrupt", repair,
	)
	corruptCuratedParitySourceSHA(t, ctx, database)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_parity_corrupt", parity[:2],
	)
	err = applyLayoutMigrationExpectErrorPG18(
		t, ctx, database, "layout_parity_corrupt", parity[2],
	)
	assertLayoutSQLState(t, err, "23514")
	assertLayoutRepairStatePG18(
		t, ctx, database, "layout_parity_corrupt",
		"organization:2,role:2,user:2", 6,
	)
	assertValidLayoutRepairPG18(
		t, ctx, database, oldRows, oldCatalogs,
	)
	assertMixedCustomLayoutRepairPG18(t, ctx, database)
	applyLayoutMigrationSetPG18(
		t, ctx, database, layoutContractSchema, parity,
	)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_a", parity,
	)
	assertCuratedParityRepairPG18(t, ctx, database)
	assertMixedCustomCuratedParityPG18(t, ctx, database)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_customer", repair,
	)
	applyLayoutMigrationSetPG18(
		t, ctx, database, "layout_customer", parity,
	)
	assertBaselineLessCustomerLayoutUntouched(
		t, ctx, database, customerSnapshot,
	)
	for _, fixture := range []struct {
		schema, generations string
	}{
		{"layout_b", "organization:1,role:1,user:1"},
		{"layout_missing", "organization:1,user:1"},
		{"layout_semantic", "organization:1,role:1,user:1"},
	} {
		applyLayoutMigrationSetPG18(
			t, ctx, database, fixture.schema, repair[:4],
		)
		err := applyLayoutMigrationExpectErrorPG18(
			t, ctx, database, fixture.schema, repair[4],
		)
		assertLayoutSQLState(t, err, "23514")
		assertLayoutRepairStatePG18(
			t, ctx, database, fixture.schema,
			fixture.generations, 3,
		)
	}
}

func assertBaselineLessCustomerLayoutUntouched(
	t *testing.T, ctx context.Context, database *sql.DB, before string,
) {
	t.Helper()
	after := layoutRepairSnapshotPG18(
		t, ctx, database, "layout_customer", "publications",
	)
	changed, err := callInstallationLayoutRepairPG18(
		ctx, database, "layout_customer",
	)
	if err != nil || changed != 0 || after != before {
		t.Fatalf("baseline-less customer layout changed=%d: %v", changed, err)
	}
}

func assertValidLayoutRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	oldRows, oldCatalogs string,
) {
	t.Helper()
	assertLayoutRepairStatePG18(
		t, ctx, database, layoutContractSchema,
		"organization:2,role:2,user:2", 6,
	)
	if got := layoutRepairSnapshotPG18(
		t, ctx, database, layoutContractSchema, "old_revisions",
	); got != oldRows {
		t.Fatal("installation predecessor revision rows changed")
	}
	if got := layoutRepairSnapshotPG18(
		t, ctx, database, layoutContractSchema, "publications",
	); got != oldCatalogs {
		t.Fatal("installation predecessor catalogs changed")
	}
	changed, err := callInstallationLayoutRepairPG18(
		ctx, database, layoutContractSchema,
	)
	if err != nil || changed != 0 {
		t.Fatalf("layout repair replay changed=%d: %v", changed, err)
	}
}

func assertMixedCustomLayoutRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	assertLayoutRepairStatePG18(
		t, ctx, database, "layout_a",
		"organization:2,role:2,user:2", 6,
	)
	var revision string
	err := database.QueryRowContext(ctx, `
SELECT revision_id::text FROM layout_a.system_resource_layout_publications
WHERE resource_kind='user'`).Scan(&revision)
	if err != nil || revision != layoutSecondRevision {
		t.Fatalf("custom user revision=%q: %v", revision, err)
	}
}

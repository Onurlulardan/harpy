//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func withDocumentsRecoveryPrincipal(t *testing.T, ctx context.Context,
	database *sql.DB, principal, schema string, action func(*sql.Tx),
) {
	t.Helper()
	if principal != "harpy_backup" && principal != "harpy_restore" {
		t.Fatal("invalid Documents recovery principal fixture")
	}
	if schema != "documents_fresh_a" && schema != "documents_adopted_a" {
		t.Fatal("invalid Documents recovery schema fixture")
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx,
		"SET LOCAL SESSION AUTHORIZATION "+principal); err != nil {
		t.Fatal(err)
	}
	if principal == "harpy_backup" {
		if _, err = tx.ExecContext(ctx,
			"SET LOCAL ROLE harpy_backup_reader"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+",pg_catalog"); err != nil {
		t.Fatal(err)
	}
	action(tx)
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func applyDocumentsRecoveryLocator(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	withDocumentsRecoveryPrincipal(t, ctx, database, "harpy_restore",
		fixture.schema, func(tx *sql.Tx) {
			for attempt := 0; attempt < 2; attempt++ {
				var changed bool
				err := tx.QueryRowContext(ctx, `SELECT
 documents_recovery_apply_locator_v1($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
					documentsRecoveryRestore, documentsRecoverySet,
					documentsRecoveryVersion, fixture.objectKey,
					fixture.objectVersion, fixture.objectKey,
					fixture.targetVersion, int64(2), fixture.ciphertextSHA256[:]).
					Scan(&changed)
				if err != nil || !changed {
					t.Fatalf("apply/replay Documents locator = %v/%v", changed, err)
				}
			}
		})
}

func terminalizeDocumentsRecoveryTransients(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	withDocumentsRecoveryPrincipal(t, ctx, database, "harpy_restore",
		fixture.schema, func(tx *sql.Tx) {
			for attempt := 0; attempt < 2; attempt++ {
				var uploads, expired, cancelled int64
				err := tx.QueryRowContext(ctx,
					"SELECT * FROM documents_recovery_terminalize_transients_v2($1)",
					documentsRecoveryRestore).
					Scan(&uploads, &expired, &cancelled)
				if err != nil || uploads != 1 || expired != 1 || cancelled != 1 {
					t.Fatalf("terminalize/replay Documents transients = %d/%d/%d/%v",
						uploads, expired, cancelled, err)
				}
			}
		})
	var upload, readyArchive, pendingArchive string
	err := database.QueryRowContext(ctx, `SELECT
 (SELECT lifecycle_state FROM `+fixture.schema+`.documents_uploads
  WHERE id='018f0000-0000-7000-8000-000000000a30'),
 (SELECT lifecycle_state FROM `+fixture.schema+`.documents_archives
  WHERE id='018f0000-0000-7000-8000-000000000a43'),
 (SELECT lifecycle_state FROM `+fixture.schema+`.documents_archives
  WHERE id='018f0000-0000-7000-8000-000000000a40')`).
		Scan(&upload, &readyArchive, &pendingArchive)
	if err != nil || upload != "CANCELLED" || readyArchive != "EXPIRED" ||
		pendingArchive != "CANCELLED" {
		t.Fatalf("Documents transient states = %s/%s/%s/%v",
			upload, readyArchive, pendingArchive, err)
	}
}

func releaseDocumentsRecoveryHold(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	withDocumentsRecoveryPrincipal(t, ctx, database, "harpy_backup",
		fixture.schema, func(tx *sql.Tx) {
			for attempt := 0; attempt < 2; attempt++ {
				var released bool
				err := tx.QueryRowContext(ctx,
					"SELECT documents_recovery_release_hold_v2($1,$2)",
					documentsRecoverySet, fixture.inventorySHA256[:]).
					Scan(&released)
				if err != nil || !released {
					t.Fatalf("release/replay Documents hold = %v/%v", released, err)
				}
			}
		})
}

func assertDocumentsRecoveryResult(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	var state, objectVersion string
	var generation, remaps int64
	err := database.QueryRowContext(ctx, `SELECT hold.lifecycle_state,
 version.object_version,version.object_generation,
 (SELECT count(*) FROM `+fixture.schema+`.documents_restore_locator_remaps
  WHERE restore_operation_id=$2)
 FROM `+fixture.schema+`.documents_dr_holds AS hold
 JOIN `+fixture.schema+`.documents_versions AS version ON version.id=$3
 WHERE hold.recovery_set_id=$1`, documentsRecoverySet,
		documentsRecoveryRestore, documentsRecoveryVersion).
		Scan(&state, &objectVersion, &generation, &remaps)
	if err != nil || state != "RELEASED" || objectVersion != fixture.targetVersion ||
		generation != 2 || remaps != 1 {
		t.Fatalf("Documents recovery result = %s/%s/%d/%d/%v",
			state, objectVersion, generation, remaps, err)
	}
}

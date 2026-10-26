//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertDocumentsRecoveryOwnerDenied(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	var state string
	err := database.QueryRowContext(ctx, "SELECT "+fixture.schema+
		".documents_recovery_begin_hold_v1($1,$2,$3,$4)",
		documentsRecoverySet, documentsRecoveryEnvironment, fixture.tenantID,
		fixture.tenantSetSHA256[:]).Scan(&state)
	requireDocumentsPGError(t, err, "42501")
}

func beginDocumentsRecoveryHold(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	withDocumentsRecoveryPrincipal(t, ctx, database, "harpy_backup",
		fixture.schema, func(tx *sql.Tx) {
			for attempt := 0; attempt < 2; attempt++ {
				var state string
				err := tx.QueryRowContext(ctx,
					"SELECT documents_recovery_begin_hold_v1($1,$2,$3,$4)",
					documentsRecoverySet, documentsRecoveryEnvironment,
					fixture.tenantID, fixture.tenantSetSHA256[:]).Scan(&state)
				if err != nil || state != "HOLDING" {
					t.Fatalf("begin/replay Documents recovery hold = %s/%v", state, err)
				}
			}
		})
}

func assertDocumentsRecoveryDeletionFence(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	withDocumentsSchema(t, ctx, database, fixture.schema, func(tx *sql.Tx) {
		execDocumentsSQL(t, ctx, tx,
			"adapters/postgres/deletion/sql/tenant/claim.sql",
			documentsRecoveryDeletion,
			"018f0000-0000-7000-8000-000000000a17",
			now, now.Add(time.Minute))
	})
	var state string
	err := database.QueryRowContext(ctx, "SELECT lifecycle_state FROM "+
		fixture.schema+".documents_blob_deletions WHERE id=$1",
		documentsRecoveryDeletion).Scan(&state)
	if err != nil || state != "PENDING" {
		t.Fatalf("Documents DR hold deletion fence = %s/%v", state, err)
	}
}

func copyAndSealDocumentsRecovery(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	withDocumentsRecoveryPrincipal(t, ctx, database, "harpy_backup",
		fixture.schema, func(tx *sql.Tx) {
			rows, err := tx.QueryContext(ctx, `SELECT version_id,object_key,
 object_version,object_generation,ciphertext_sha256,ciphertext_bytes,key_id
 FROM documents_recovery_inventory_v1($1,$2)`,
				documentsRecoverySet, fixture.tenantID)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			assertDocumentsRecoveryInventoryRow(t, rows, fixture)

			if _, err = tx.ExecContext(ctx, "SAVEPOINT wrong_copy"); err != nil {
				t.Fatal(err)
			}
			wrong := [32]byte{1}
			_, err = tx.ExecContext(ctx,
				"SELECT documents_recovery_record_copy_v1($1,$2,$3,$4,$5,$6,$7,$8,$9)",
				documentsRecoverySet, documentsRecoveryVersion, fixture.objectKey,
				fixture.objectVersion, documentsRecoveryTargetKey(fixture),
				"recovery-copy-version-1", wrong[:], int64(117), "documents-kek-1")
			requireDocumentsPGError(t, err, "42501")
			if _, err = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT wrong_copy"); err != nil {
				t.Fatal(err)
			}
			for attempt := 0; attempt < 2; attempt++ {
				_, err = tx.ExecContext(ctx,
					"SELECT documents_recovery_record_copy_v1($1,$2,$3,$4,$5,$6,$7,$8,$9)",
					documentsRecoverySet, documentsRecoveryVersion, fixture.objectKey,
					fixture.objectVersion, documentsRecoveryTargetKey(fixture),
					"recovery-copy-version-1", fixture.ciphertextSHA256[:],
					int64(117), "documents-kek-1")
				if err != nil {
					t.Fatalf("record/replay Documents recovery copy: %v", err)
				}
			}
			for attempt := 0; attempt < 2; attempt++ {
				var sealed bool
				err = tx.QueryRowContext(ctx,
					"SELECT documents_recovery_seal_hold_v2($1,$2,$3)",
					documentsRecoverySet, fixture.inventorySHA256[:], int64(1)).
					Scan(&sealed)
				if err != nil || !sealed {
					t.Fatalf("seal/replay Documents recovery hold = %v/%v", sealed, err)
				}
			}
		})
}

func assertDocumentsRecoveryInventoryRow(t *testing.T, rows *sql.Rows,
	fixture documentsRecoveryFixture,
) {
	t.Helper()
	var version, key, objectVersion, keyID string
	var generation, bytes int64
	var digest []byte
	if !rows.Next() {
		t.Fatal("Documents recovery inventory is empty")
	}
	err := rows.Scan(&version, &key, &objectVersion, &generation,
		&digest, &bytes, &keyID)
	if err != nil || version != documentsRecoveryVersion || key != fixture.objectKey ||
		objectVersion != fixture.objectVersion || generation != 1 || bytes != 117 ||
		keyID != "documents-kek-1" || string(digest) != string(fixture.ciphertextSHA256[:]) {
		t.Fatalf("Documents recovery inventory differs: %s/%s/%s/%d/%d/%s/%v",
			version, key, objectVersion, generation, bytes, keyID, err)
	}
	if rows.Next() || rows.Err() != nil {
		t.Fatalf("Documents recovery inventory is not exact: %v", rows.Err())
	}
}

func documentsRecoveryTargetKey(fixture documentsRecoveryFixture) string {
	return "environments/" + documentsRecoveryEnvironment + "/recovery-sets/" +
		documentsRecoverySet + "/documents/tenants/" + fixture.tenantID +
		"/versions/" + documentsRecoveryVersion + ".fcdoc1"
}

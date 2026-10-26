//go:build integration

package maintenance

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertMaintenancePG18EvidenceConstraints(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const insert = `
INSERT INTO t_iam_a.iam_directory_key_evidence(
 lookup_key_id,last_job_id,projected_row_count,zero_rows_since,
 last_result_sha256,revision,updated_at)
VALUES($1,$2::uuid,1,NULL,$3::bytea,1,pg_catalog.transaction_timestamp())`
	digest := bytes.Repeat([]byte{0x41}, 32)
	for _, keyID := range []string{iamMaintenanceKeyID, "legacy-key-v1"} {
		if _, err := database.ExecContext(
			ctx, insert, keyID, iamMaintenanceJobID, digest,
		); err != nil {
			t.Fatalf("insert key evidence %q: %v", keyID, err)
		}
	}
	if _, err := database.ExecContext(
		ctx, insert, "invalid*key", iamMaintenanceJobID, digest,
	); err == nil {
		t.Fatal("invalid maintenance key evidence passed")
	}
}

func assertMaintenancePG18Isolation(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	const statement = `
SELECT
 (SELECT pg_catalog.count(*) FROM t_iam_a.iam_maintenance_schedules),
 (SELECT pg_catalog.count(*) FROM t_iam_b.iam_maintenance_schedules),
 (SELECT pg_catalog.count(*) FROM t_iam_a.iam_directory_key_evidence),
 (SELECT pg_catalog.count(*) FROM t_iam_b.iam_directory_key_evidence),
 (SELECT pg_catalog.count(*) FROM platform.iam_login_directory)`
	var schedulesA, schedulesB, evidenceA, evidenceB, directory int
	err := connection.QueryRowContext(ctx, statement).Scan(
		&schedulesA, &schedulesB, &evidenceA, &evidenceB, &directory)
	if err != nil || schedulesA != 1 || schedulesB != 0 ||
		evidenceA != 2 || evidenceB != 0 || directory != 0 {
		t.Fatalf("isolation schedules=%d/%d evidence=%d/%d directory=%d: %v",
			schedulesA, schedulesB, evidenceA, evidenceB, directory, err)
	}
}

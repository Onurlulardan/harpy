//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	cutoverapplication "github.com/Onurlulardan/harpy/core/internal/platform/cutover/application"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestFullRolloutCompletionAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			skipped := map[string]bool{
				"20260716020103_seed_global_billing_authority." + origin + ".sql": true,
			}
			database := startCutoverIntegrationDatabaseForOriginSkipping(
				t, ctx, origin, skipped,
			)
			environmentID := uuid.MustParse(integrationEnvironment)
			tenantID := uuid.MustParse(integrationTenantID)
			receiptID := uuid.MustParse("50000000-0000-4000-8000-000000000001")
			now := time.Now().UTC().Truncate(time.Microsecond)
			seedFullRolloutAuthority(t, ctx, database.admin, environmentID, tenantID)
			assertFullRolloutPrivileges(t, ctx, database.admin)
			setClusterAdminLease(t, ctx, database.admin, now.Add(10*time.Minute))
			connection := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
			defer connection.Close(context.Background())
			recorder, err := NewFullRolloutRecorder(connection)
			if err != nil {
				t.Fatal(err)
			}
			value := fullRolloutCompletionFixture(environmentID, tenantID, receiptID, now)
			earlyClosure := value
			earlyClosure.Evidence.RollbackWindowClosedAt =
				earlyClosure.Evidence.LastLegacyHitAt.Add(-time.Microsecond)
			_, err = recorder.Record(ctx, earlyClosure, [32]byte{1})
			requireFullRolloutSQLState(t, err, "P57F0")
			_, err = database.admin.Exec(ctx, `INSERT INTO platform.contract_retirements (
contract_key,owner_module,domain,state,remaining_scope_count,hit_count,revision,updated_at
) VALUES ('phase13.blocker.v1','work','work','active',0,0,1,clock_timestamp())`)
			if err != nil {
				t.Fatal(err)
			}
			_, err = cutoverapplication.RecordFullRollout(ctx, recorder, value)
			requireFullRolloutSQLState(t, err, "P57F4")
			_, err = database.admin.Exec(ctx, `WITH moment AS (
SELECT pg_catalog.clock_timestamp() AS value)
UPDATE platform.contract_retirements SET state='eligible',remaining_scope_count=0,
zero_hit_since=moment.value-interval '31 days',eligible_at=moment.value,
revision=revision+1,updated_at=moment.value FROM moment
WHERE contract_key='phase13.blocker.v1';
WITH moment AS (SELECT pg_catalog.clock_timestamp() AS value)
UPDATE platform.contract_retirements SET state='retired',retired_at=moment.value,
revision=revision+1,updated_at=moment.value FROM moment
WHERE contract_key='phase13.blocker.v1'`, pgx.QueryExecModeSimpleProtocol)
			if err != nil {
				t.Fatal(err)
			}
			receipt, err := cutoverapplication.RecordFullRollout(ctx, recorder, value)
			if err != nil {
				t.Fatal(err)
			}
			digest, err := receipt.SHA256()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := cutoverapplication.RecordFullRollout(ctx, recorder, value); err != nil {
				t.Fatalf("exact full rollout replay: %v", err)
			}
			tampered := value
			tampered.Evidence.RollbackWindowClosureSHA256[0] ^= 0xff
			_, err = recorder.Record(ctx, tampered, digest)
			requireFullRolloutSQLState(t, err, "P57F5")
			var persisted, rollbackClosure string
			err = database.admin.QueryRow(ctx, `SELECT pg_catalog.encode(evidence_sha256,'hex'),
pg_catalog.encode(rollback_window_closure_sha256,'hex')
FROM cutover_authority.full_rollout_completion_receipts WHERE receipt_id=$1`,
				receiptID).Scan(&persisted, &rollbackClosure)
			if err != nil || persisted != encodeDigest(digest) ||
				rollbackClosure != encodeDigest(value.Evidence.RollbackWindowClosureSHA256) {
				t.Fatalf("persisted full rollout digests = %s/%s, %v",
					persisted, rollbackClosure, err)
			}

			other := fullRolloutCompletionFixture(environmentID,
				uuid.MustParse(integrationTenantBID),
				uuid.MustParse("50000000-0000-4000-8000-000000000002"), now)
			_, err = cutoverapplication.RecordFullRollout(ctx, recorder, other)
			requireFullRolloutSQLState(t, err, "P57F1")
			_, err = connection.Exec(ctx, `UPDATE cutover_authority.full_rollout_completion_receipts
SET v1_writers=1 WHERE receipt_id=$1`, receiptID)
			requireFullRolloutSQLState(t, err, "42501")
			_, err = database.admin.Exec(ctx, `DELETE FROM
cutover_authority.full_rollout_completion_receipts WHERE receipt_id=$1`, receiptID)
			requireFullRolloutSQLState(t, err, "55000")
		})
	}
}

func requireFullRolloutSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("full rollout error = %v, want SQLSTATE %s", err, want)
	}
}

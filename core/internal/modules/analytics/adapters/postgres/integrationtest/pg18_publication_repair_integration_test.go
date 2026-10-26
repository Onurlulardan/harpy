//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	analyticsRepairRelease = "019b1000-0000-7000-8000-000000000031"
	analyticsRepairVersion = "019b1000-0000-7000-8000-000000000032"
	analyticsRepairActor   = "019b1000-0000-7000-8000-000000000033"
)

func assertAnalyticsPublicationRepair(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_publication_releases (
id,resource_kind,resource_version_id,audience,lifecycle_state,cache_policy,
manifest_sha256,requester_id,approver_id,row_revision,generation,public_id,
created_at,updated_at,submitted_at,approved_at,activated_at
) VALUES ($1,'query',$2,'public','active','none',decode(repeat('31',32),'hex'),
$3,'019b1000-0000-7000-8000-000000000034',4,1,decode(repeat('42',16),'hex'),
clock_timestamp()-interval '1 hour',clock_timestamp()-interval '30 minutes',
clock_timestamp()-interval '50 minutes',clock_timestamp()-interval '40 minutes',
clock_timestamp()-interval '30 minutes')`, analyticsRepairRelease,
		analyticsRepairVersion, analyticsRepairActor)
	if err != nil {
		t.Fatalf("prepare Analytics repair release: %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			`RESET ROLE; RESET SESSION AUTHORIZATION`)
		_ = connection.Close()
	}()
	if _, err = connection.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_v2_runtime; SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	callAnalyticsRepair(t, ctx, connection, 1)
	assertAnalyticsPointerState(t, ctx, database, "active", 1)
	if _, err = database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_publication_releases
SET lifecycle_state='revoked',row_revision=5,generation=2,
revoked_at=clock_timestamp(),updated_at=clock_timestamp()
WHERE id=$1`, analyticsRepairRelease); err != nil {
		t.Fatal(err)
	}
	callAnalyticsRepair(t, ctx, connection, 2)
	assertAnalyticsPointerState(t, ctx, database, "revoked", 2)
}

func callAnalyticsRepair(
	t *testing.T, ctx context.Context, connection *sql.Conn, generation int,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err == nil {
		var repaired bool
		err = transaction.QueryRowContext(ctx,
			`SELECT analytics_publication_repair_v1('t_analytics_a',$1,$2,clock_timestamp())`,
			analyticsRepairRelease, generation).Scan(&repaired)
		if err == nil && !repaired {
			err = sql.ErrNoRows
		}
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("repair Analytics publication generation %d: %v", generation, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsPointerState(
	t *testing.T, ctx context.Context, database *sql.DB, state string, generation int,
) {
	t.Helper()
	var actualState string
	var actualGeneration int
	err := database.QueryRowContext(ctx, `SELECT state,generation
FROM platform.analytics_publications WHERE release_id=$1`,
		analyticsRepairRelease).Scan(&actualState, &actualGeneration)
	if err != nil || actualState != state || actualGeneration != generation {
		t.Fatalf("Analytics pointer=%s/%d err=%v", actualState, actualGeneration, err)
	}
}

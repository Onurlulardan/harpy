//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	analyticsLimiterRelease = "019b1000-0000-7000-8000-000000000041"
	analyticsLimiterVersion = "019b1000-0000-7000-8000-000000000042"
)

func assertAnalyticsPublicLimiter(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.analytics_publications(
public_id,tenant_id,release_id,resource_kind,resource_version_id,audience,state,
generation,capability_sha256,capability_expires_at,activated_at,revoked_at)
VALUES(decode(repeat('51',16),'hex'),$1,$2,'query',$3,'public','active',1,
NULL,NULL,$4,NULL)`, analyticsTenantA, analyticsLimiterRelease,
		analyticsLimiterVersion, analyticsLimiterTime())
	if err != nil {
		t.Fatalf("prepare Analytics limiter release: %v", err)
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
	leases := make([]string, 0, 11)
	for index := 0; index < 10; index++ {
		leases = append(leases, admitAnalyticsPublic(t, ctx, connection, 1))
	}
	if _, err = connection.ExecContext(ctx, `SELECT platform.analytics_public_admit_v1(
$1,$2,1,'203.0.113.10',$3)`, analyticsTenantA, analyticsLimiterRelease,
		analyticsLimiterTime()); err == nil {
		t.Fatal("Analytics public concurrency limit admitted request eleven")
	} else {
		requireAnalyticsPGCode(t, err, "P6105", "public concurrency limit")
	}
	releaseAnalyticsPublic(t, ctx, connection, leases[0])
	leases = append(leases[1:], admitAnalyticsPublic(t, ctx, connection, 1))
	if _, err = connection.ExecContext(ctx, `SELECT platform.analytics_public_admit_v1(
$1,$2,2,'203.0.113.11',$3)`, analyticsTenantA, analyticsLimiterRelease,
		analyticsLimiterTime()); err == nil {
		t.Fatal("Analytics public limiter accepted stale generation")
	} else {
		requireAnalyticsPGCode(t, err, "P6102", "stale publication generation")
	}
	if _, err = connection.ExecContext(ctx,
		`SELECT count(*) FROM platform.analytics_public_concurrency_leases`); err == nil {
		t.Fatal("Analytics runtime read limiter authority tables directly")
	} else {
		requireAnalyticsPGCode(t, err, "42501", "limiter table read")
	}
	for _, lease := range leases {
		releaseAnalyticsPublic(t, ctx, connection, lease)
	}
}

func admitAnalyticsPublic(
	t *testing.T, ctx context.Context, connection *sql.Conn, generation int,
) string {
	t.Helper()
	var lease string
	err := connection.QueryRowContext(ctx, `SELECT platform.analytics_public_admit_v1(
$1,$2,$3,'203.0.113.10',$4)`, analyticsTenantA, analyticsLimiterRelease,
		generation, analyticsLimiterTime()).Scan(&lease)
	if err != nil || lease == "" {
		t.Fatalf("admit Analytics public request: lease=%q err=%v", lease, err)
	}
	return lease
}

func releaseAnalyticsPublic(
	t *testing.T, ctx context.Context, connection *sql.Conn, lease string,
) {
	t.Helper()
	_, err := connection.ExecContext(ctx, `SELECT platform.analytics_public_release_v1(
$1,$2,$3,1)`, lease, analyticsTenantA, analyticsLimiterRelease)
	if err != nil {
		t.Fatalf("release Analytics public lease: %v", err)
	}
}

func analyticsLimiterTime() time.Time {
	return time.Date(2026, 7, 18, 16, 15, 0, 123000, time.UTC)
}

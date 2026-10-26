//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const analyticsLegacyEndpoint = "11111111-1111-4111-8111-111111111141"

func assertAnalyticsLegacyRetirement(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.analytics_legacy_public_routes(
legacy_tenant_identity,method,path,slug,tenant_id,legacy_endpoint_id,public_id,
release_id,state,hit_count,last_hit_at,consumers_acked_at,remove_after)
VALUES('tenant-a','GET','/api/openapi/execute/orders','orders',$1,$2,
decode(repeat('51',16),'hex'),$3,'active',0,NULL,NULL,$4)`, analyticsTenantA,
		analyticsLegacyEndpoint, analyticsLimiterRelease,
		analyticsLegacyRetirementTime().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("prepare Analytics legacy mapping: %v", err)
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
	var release string
	err = connection.QueryRowContext(ctx, `SELECT release_id::text
FROM platform.analytics_legacy_public_resolve_v1(
'tenant-a','GET','/api/openapi/execute/orders','orders')`).Scan(&release)
	if err != nil || release != analyticsLimiterRelease {
		t.Fatalf("resolve Analytics legacy mapping: release=%q err=%v", release, err)
	}
	if _, err = connection.ExecContext(ctx, `SELECT *
FROM platform.analytics_legacy_public_resolve_v1(
'tenant-b','GET','/api/openapi/execute/orders','orders')`); err == nil {
		t.Fatal("Analytics legacy resolver accepted tenant substitution")
	} else {
		requireAnalyticsPGCode(t, err, "P6102", "legacy tuple substitution")
	}
	var hitCount int
	err = connection.QueryRowContext(ctx, `SELECT hit_count
FROM platform.analytics_legacy_public_record_hit_v1(
'tenant-a','GET','/api/openapi/execute/orders','orders',$1,$2,
decode(repeat('51',16),'hex'),$3,$4)`, analyticsTenantA, analyticsLegacyEndpoint,
		analyticsLimiterRelease, analyticsLegacyHitTime()).Scan(&hitCount)
	if err != nil || hitCount != 1 {
		t.Fatalf("record Analytics legacy hit: count=%d err=%v", hitCount, err)
	}
	if _, err = connection.ExecContext(ctx, `RESET ROLE; RESET SESSION AUTHORIZATION;
SET SESSION AUTHORIZATION harpy_cluster_admin`); err != nil {
		t.Fatal(err)
	}
	ackAnalyticsLegacyConsumers(t, ctx, connection)
	retireAnalyticsLegacyMapping(t, ctx, connection)
	if _, err = connection.ExecContext(ctx, `RESET SESSION AUTHORIZATION;
SET SESSION AUTHORIZATION harpy_v2_runtime; SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx, `SELECT *
FROM platform.analytics_legacy_public_resolve_v1(
'tenant-a','GET','/api/openapi/execute/orders','orders')`); err == nil {
		t.Fatal("Analytics legacy resolver served retired mapping")
	} else {
		requireAnalyticsPGCode(t, err, "P6102", "retired legacy mapping")
	}
}

func ackAnalyticsLegacyConsumers(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	var state string
	err := connection.QueryRowContext(ctx, `SELECT state
FROM platform.analytics_legacy_public_ack_consumers_v1(
'tenant-a','GET','/api/openapi/execute/orders','orders',$1,$2,
decode(repeat('51',16),'hex'),$3,decode(repeat('61',32),'hex'),$4)`,
		analyticsTenantA, analyticsLegacyEndpoint, analyticsLimiterRelease,
		analyticsLegacyRetirementTime().Add(-time.Hour)).Scan(&state)
	if err != nil || state != "active" {
		t.Fatalf("ack Analytics legacy consumers: state=%q err=%v", state, err)
	}
}

func retireAnalyticsLegacyMapping(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	at := analyticsLegacyRetirementTime()
	var state string
	err := connection.QueryRowContext(ctx, `SELECT state
FROM platform.analytics_legacy_public_retire_v1(
'tenant-a','GET','/api/openapi/execute/orders','orders',$1,$2,
decode(repeat('51',16),'hex'),$3,decode(repeat('62',32),'hex'),$4,$5)`,
		analyticsTenantA, analyticsLegacyEndpoint, analyticsLimiterRelease,
		at.Add(-time.Minute), at).Scan(&state)
	if err != nil || state != "retired" {
		t.Fatalf("retire Analytics legacy mapping: state=%q err=%v", state, err)
	}
}

func analyticsLegacyHitTime() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 123000, time.UTC)
}

func analyticsLegacyRetirementTime() time.Time {
	return time.Date(2026, 7, 18, 17, 0, 0, 123000, time.UTC)
}

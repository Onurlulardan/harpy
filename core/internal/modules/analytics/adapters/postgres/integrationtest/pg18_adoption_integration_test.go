//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	analyticsAdoptionAuthority = "019b1000-0000-7000-8000-000000000031"
	analyticsAdoptionRun       = "11111111-1111-4111-8111-111111111131"
	analyticsLegacyAdopted     = "11111111-1111-4111-8111-111111111132"
	analyticsLegacyRetired     = "11111111-1111-4111-8111-111111111133"
	analyticsAdoptedQuery      = "019b1000-0000-7000-8000-000000000034"
	analyticsAdoptedVersion    = "019b1000-0000-7000-8000-000000000035"
	analyticsAdoptedRelease    = "019b1000-0000-7000-8000-000000000036"
)

func assertAnalyticsV1Adoption(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(), `RESET SESSION AUTHORIZATION`)
		_ = connection.Close()
	}()
	if _, err = connection.ExecContext(ctx, `SET SESSION AUTHORIZATION harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx, `SELECT platform.analytics_register_tenant_authority_v1(
$1,$2,2,2,$3)`, analyticsAdoptionAuthority, analyticsTenantA, analyticsAdoptionTime()); err == nil {
		t.Fatal("legacy Analytics authority helper remained executable")
	} else {
		requireAnalyticsPGCode(t, err, "42501", "legacy authority helper")
	}
	if _, err = connection.ExecContext(ctx,
		`SET search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	_, err = connection.ExecContext(ctx, `SELECT * FROM analytics_adoption_record_item_v1(
$1::name,$2,$3,$4,$5,decode(repeat('11',32),'hex'),'retired','RETIRED',
NULL,NULL,NULL,NULL,$6)`, "t_analytics_a", analyticsAdoptionAuthority,
		analyticsTenantA, analyticsAdoptionRun, analyticsLegacyRetired, analyticsAdoptionTime())
	if err == nil {
		t.Fatal("legacy Analytics item helper remained executable")
	}
	requireAnalyticsPGCode(t, err, "42501", "legacy item helper")
	if _, err = connection.ExecContext(ctx,
		`SELECT count(*) FROM analytics_v1_adoption_items`); err == nil {
		t.Fatal("Analytics adopter read owner table directly")
	} else {
		requireAnalyticsPGCode(t, err, "42501", "adopter direct read")
	}
	_, err = connection.ExecContext(ctx, `SELECT * FROM analytics_adoption_accept_v1(
$1::name,$2,$3,$4,2,0,0,0,decode(repeat('11',32),'hex'),$5)`,
		"t_analytics_a", analyticsAdoptionAuthority, analyticsTenantA,
		analyticsAdoptionRun, analyticsAdoptionTime())
	if err == nil {
		t.Fatal("legacy Analytics acceptance helper remained executable")
	}
	requireAnalyticsPGCode(t, err, "42501", "legacy acceptance helper")
}

func setAnalyticsShadowAuthority(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `WITH changed AS (
    UPDATE platform.cutover_groups SET generation=2,state='shadow',writer='v1'
    WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'
    RETURNING scope_kind,scope_id
)
UPDATE platform.domain_cutovers AS domain_row SET generation=2,state='shadow',writer='v1'
FROM changed WHERE domain_row.scope_kind=changed.scope_kind
  AND domain_row.scope_id=changed.scope_id AND domain_row.domain='analytics'`, analyticsTenantA)
	if err != nil {
		t.Fatalf("set Analytics shadow authority: %v", err)
	}
}

func recordAnalyticsAdoptionItems(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	at := analyticsAdoptionTime()
	var replay bool
	err := connection.QueryRowContext(ctx, `SELECT replay FROM analytics_adoption_record_item_v1(
$1::name,$2,$3,$4,$5,decode(repeat('11',32),'hex'),'adopted',NULL,
$6,$7,$8,decode(repeat('21',32),'hex'),$9)`, "t_analytics_a", analyticsAdoptionAuthority,
		analyticsTenantA, analyticsAdoptionRun, analyticsLegacyAdopted, analyticsAdoptedQuery,
		analyticsAdoptedVersion, analyticsAdoptedRelease, at).Scan(&replay)
	if err != nil || replay {
		t.Fatalf("record adopted Analytics endpoint: replay=%v err=%v", replay, err)
	}
	err = connection.QueryRowContext(ctx, `SELECT replay FROM analytics_adoption_record_item_v1(
$1::name,$2,$3,$4,$5,decode(repeat('12',32),'hex'),'retired','UNSUPPORTED_WINDOW',
NULL,NULL,NULL,NULL,$6)`, "t_analytics_a", analyticsAdoptionAuthority, analyticsTenantA,
		analyticsAdoptionRun, analyticsLegacyRetired, at).Scan(&replay)
	if err != nil || replay {
		t.Fatalf("record retired Analytics endpoint: replay=%v err=%v", replay, err)
	}
}

func acceptAnalyticsAdoption(
	ctx context.Context, connection *sql.Conn, root [32]byte,
) (bool, error) {
	var replay bool
	err := connection.QueryRowContext(ctx, `SELECT replay FROM analytics_adoption_accept_v1(
$1::name,$2,$3,$4,2,2,1,1,$5,$6)`, "t_analytics_a", analyticsAdoptionAuthority,
		analyticsTenantA, analyticsAdoptionRun, root[:], analyticsAdoptionTime()).Scan(&replay)
	return replay, err
}

func analyticsAdoptionRoot() [32]byte {
	adopted := strings.Join([]string{analyticsLegacyAdopted, strings.Repeat("11", 32),
		"adopted", "", analyticsAdoptedQuery, analyticsAdoptedVersion,
		analyticsAdoptedRelease, strings.Repeat("21", 32)}, "|")
	retired := strings.Join([]string{analyticsLegacyRetired, strings.Repeat("12", 32),
		"retired", "UNSUPPORTED_WINDOW", "", "", "", ""}, "|")
	return sha256.Sum256([]byte(adopted + "\n" + retired))
}

func assertAnalyticsReverseEvidence(
	t *testing.T, ctx context.Context, connection *sql.Conn, root [32]byte,
) {
	t.Helper()
	var count int
	var stored []byte
	err := connection.QueryRowContext(ctx, `SELECT source_count,root_sha256
FROM t_analytics_a.analytics_v1_adoption_acceptance_receipts WHERE run_id=$1`,
		analyticsAdoptionRun).Scan(&count, &stored)
	if err != nil || count != 2 || hex.EncodeToString(stored) != hex.EncodeToString(root[:]) {
		t.Fatalf("Analytics reverse evidence count=%d root=%x err=%v", count, stored, err)
	}
	_, err = connection.ExecContext(ctx, `UPDATE t_analytics_a.analytics_v1_adoption_items
SET created_at=created_at WHERE run_id=$1`, analyticsAdoptionRun)
	if err == nil {
		t.Fatal("Analytics adoption evidence was mutable")
	}
	requireAnalyticsPGCode(t, err, "55000", "adoption evidence mutation")
}

func requireAnalyticsPGCode(t *testing.T, err error, want, operation string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("%s error=%v, want SQLSTATE %s", operation, err, want)
	}
}

func analyticsAdoptionTime() time.Time {
	return time.Date(2026, 7, 18, 16, 30, 0, 123000, time.UTC)
}

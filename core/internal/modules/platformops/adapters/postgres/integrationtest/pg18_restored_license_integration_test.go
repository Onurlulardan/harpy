//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	"github.com/jackc/pgx/v5/pgconn"
)

func exerciseRestoredLicenseGrant(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, receiptSHA256 string,
) {
	t.Helper()
	connection := platformOpsLicenseConnection(t, ctx, database)
	arguments := []any{
		operationID, environmentID, "abababab-abab-4bab-8bab-abababababab",
		receiptSHA256, "bcbcbcbc-bcbc-4cbc-8cbc-bcbcbcbcbcbc",
		"cdcdcdcd-cdcd-4dcd-8dcd-cdcdcdcdcdcd", "connected", "forgeguard",
		int64(4), strings.Repeat("c", 64), int64(5), strings.Repeat("d", 64),
		restoredlicensev1.ManifestSHA256(), int64(7),
	}
	first := readRestoredLicenseGrant(t, ctx, connection, arguments...)
	replay := readRestoredLicenseGrant(t, ctx, connection, arguments...)
	if first.bindingID == "" || first.bindingID != replay.bindingID ||
		first.revision != 7 || !first.leaseUntil.Valid || !first.databaseNow.Valid ||
		!first.leaseUntil.Time.After(first.databaseNow.Time) {
		t.Fatalf("restored license grant drifted: first=%+v replay=%+v", first, replay)
	}
	conflict := append([]any(nil), arguments...)
	conflict[11] = strings.Repeat("e", 64)
	if _, err := queryRestoredLicenseGrant(ctx, connection, conflict...); pgCode(err) != "23505" {
		t.Fatalf("restored license conflict code=%q err=%v", pgCode(err), err)
	}
	var count int
	err := connection.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.platformops_restored_license_grants`).Scan(&count)
	if pgCode(err) != "42501" {
		t.Fatalf("restored license table read code=%q err=%v", pgCode(err), err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='saas',maximum_customer_tenants=NULL`); err != nil {
		t.Fatal(err)
	}
	if _, err = queryRestoredLicenseGrant(ctx, connection, arguments...); pgCode(err) != "42501" {
		t.Fatalf("SaaS restored license code=%q err=%v", pgCode(err), err)
	}
	if _, err = database.ExecContext(ctx, `UPDATE platform.tenancy_deployment_policy
SET deployment_mode='onpremise',maximum_customer_tenants=0`); err != nil {
		t.Fatal(err)
	}
}

type restoredLicenseGrantRow struct {
	bindingID               string
	revision                int64
	leaseUntil, databaseNow sql.NullTime
}

func readRestoredLicenseGrant(
	t *testing.T, ctx context.Context, connection *sql.Conn, arguments ...any,
) restoredLicenseGrantRow {
	t.Helper()
	row, err := queryRestoredLicenseGrant(ctx, connection, arguments...)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

func queryRestoredLicenseGrant(
	ctx context.Context, connection *sql.Conn, arguments ...any,
) (restoredLicenseGrantRow, error) {
	var ignored [14]any
	var result restoredLicenseGrantRow
	destinations := []any{
		&ignored[0], &ignored[1], &ignored[2], &ignored[3],
		&result.bindingID, &ignored[4], &ignored[5], &ignored[6], &ignored[7],
		&ignored[8], &ignored[9], &ignored[10], &ignored[11], &ignored[12],
		&result.revision, &result.leaseUntil, &result.databaseNow,
	}
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.platformops_restored_license_scope_mint_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, arguments...).Scan(destinations...)
	return result, err
}

func platformOpsLicenseConnection(
	t *testing.T, ctx context.Context, database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = connection.ExecContext(
			context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION; RESET application_name",
		)
		_ = connection.Close()
	})
	for _, statement := range []string{
		"SET SESSION AUTHORIZATION harpy_v2_runtime",
		"SET ROLE harpy_runtime",
		"SET application_name='harpy-v2:core-operator-license-reconcile'",
	} {
		if _, err = connection.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	return connection
}

func pgCode(err error) string {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code
	}
	return ""
}

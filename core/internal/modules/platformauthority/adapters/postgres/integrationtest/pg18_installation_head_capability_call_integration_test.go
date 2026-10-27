//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func callInstallationHeadPin(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	routine, applicationName string,
	fixture authorityHeadFixture,
	capability *authorityActivationCapability,
) (bool, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if capability != nil {
		if err = mintAuthorityActivationStep(
			ctx, transaction, *capability,
		); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(
		ctx, "SET ROLE harpy_runtime",
	); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(
		ctx, "SELECT pg_catalog.set_config('application_name',$1,true)",
		applicationName,
	); err != nil {
		t.Fatal(err)
	}
	query := fmt.Sprintf(`SELECT platform.%s(
 $1,$2,$3,$4::name,$5,$6,$7,$8,$9,$10,$11,$12,$13
)`, routine)
	var replayed bool
	err = transaction.QueryRowContext(
		ctx, query,
		fixture.activationOperationID, fixture.environmentID,
		fixture.tenantID, fixture.schemaName, fixture.bindingID,
		"tenant", fixture.tenantID, int64(1), nil,
		fixture.stateSHA256, fixture.proposalSHA256,
		fixture.ackSHA256, fixture.manifestSHA256,
	).Scan(&replayed)
	if err != nil {
		_ = transaction.Rollback()
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
		return false, err
	}
	if _, err = transaction.ExecContext(
		ctx, "RESET ROLE; RESET SESSION AUTHORIZATION",
	); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		return false, err
	}
	return replayed, nil
}

func readAuthorityHeadState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture authorityHeadFixture,
) authorityHeadState {
	t.Helper()
	var state authorityHeadState
	if err := database.QueryRowContext(ctx, `
SELECT operation.status,
 (SELECT count(*) FROM platform.platformauthority_installation_head_pins
  WHERE operation_id=$2),
 (SELECT count(*) FROM platform.platformauthority_heads
  WHERE binding_id=$3 AND scope_kind='tenant' AND scope_id=$4),
 (SELECT count(*) FROM platform.platformauthority_receipts
  WHERE operation_id=$1)
FROM platform.platformauthority_operations AS operation
WHERE operation.operation_id=$1`,
		fixture.authorityOperationID, fixture.activationOperationID,
		fixture.bindingID, fixture.tenantID,
	).Scan(&state.status, &state.pins, &state.heads, &state.receipts); err != nil {
		t.Fatal(err)
	}
	return state
}

func requireAuthorityHeadState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture authorityHeadFixture,
	want authorityHeadState,
) {
	t.Helper()
	got := readAuthorityHeadState(t, ctx, database, fixture)
	if got != want {
		t.Fatalf("authority head state=%#v want=%#v", got, want)
	}
}

func authorityPGCode(err error) string {
	var pgError *pgconn.PgError
	if err != nil && errors.As(err, &pgError) {
		return pgError.Code
	}
	return ""
}

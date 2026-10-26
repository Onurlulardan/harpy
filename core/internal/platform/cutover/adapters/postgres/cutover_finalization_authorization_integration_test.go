//go:build integration

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func authorizeFinalization(
	ctx context.Context, connection *pgx.Conn, evidence finalizationEvidence,
) error {
	_, err := connection.Exec(ctx, `SELECT cutover_authority.authorize_finalization_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::text,$5::bytea,$6::text,$7::bytea,
    $8::text,$9::bytea,$10::text,$11::bytea,$12::text,$13::bytea,
    $14::text,$15::text,$16::bytea)`,
		evidence.operationID, evidence.runID, evidence.leaseID,
		evidence.leaseSHA256, []byte(evidence.leaseDocument),
		evidence.adoptionSHA256, []byte(evidence.adoptionDocument),
		evidence.sealSHA256, []byte(evidence.sealDocument),
		evidence.cleanupSHA256, []byte(evidence.cleanupDocument),
		evidence.authoritySHA256, []byte(evidence.authorityDocument),
		evidence.terminalState, evidence.terminalSHA256,
		[]byte(evidence.terminalDocument),
	)
	return err
}

func setClusterAdminLease(
	t testingT, ctx context.Context, admin *pgx.Conn, expiresAt time.Time,
) {
	t.Helper()
	statement := fmt.Sprintf(`ALTER ROLE harpy_cluster_admin LOGIN NOINHERIT
NOSUPERUSER CREATEDB CREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD 'integration-leased-secret' VALID UNTIL '%s'`, expiresAt.Format(time.RFC3339Nano))
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("set cluster admin lease: %v", err)
	}
}

func setFinalizerRoleLease(
	t testingT, ctx context.Context, admin *pgx.Conn, expiresAt time.Time,
) {
	t.Helper()
	statement := fmt.Sprintf(`ALTER ROLE harpy_cutover_finalizer LOGIN NOINHERIT
NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1
PASSWORD 'integration-leased-secret' VALID UNTIL '%s'`, expiresAt.Format(time.RFC3339Nano))
	if _, err := admin.Exec(ctx, statement); err != nil {
		t.Fatalf("set finalizer role lease: %v", err)
	}
}

type testingT interface {
	Helper()
	Fatalf(string, ...any)
}

//go:build integration

package v1adoption

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestNotificationsPG18SignedCutoverCompositeAndDeniedLegacySurfaces(t *testing.T) {
	ctx := context.Background()
	db := startNotificationPG18(t, ctx)
	prepareNotificationDatabase(t, ctx, db)
	prepareNotificationCutoverPlatform(t, ctx, db)
	fixture := prepareNotificationCutoverFixture(t, ctx, db)

	applied, keySHA, stateSHA := executeNotificationComposite(t, ctx, db, fixture)
	if applied != 4 || len(keySHA) != 64 || keySHA != stateSHA {
		t.Fatalf("composite result=%d %q %q", applied, keySHA, stateSHA)
	}
	replayed, replayKey, replayState := executeNotificationComposite(t, ctx, db, fixture)
	if replayed != applied || replayKey != keySHA || replayState != stateSHA {
		t.Fatalf("replay result=%d %q %q", replayed, replayKey, replayState)
	}
	assertNotificationCutoverProjection(t, ctx, db, fixture, stateSHA)
	assertNotificationCutoverTamperDenied(t, ctx, db, fixture)
	assertNotificationLegacySurfacesDenied(t, ctx, db, fixture.tenantID)
	assertNotificationCutoverReceiptImmutable(t, ctx, db, fixture)
}

func executeNotificationComposite(t *testing.T, ctx context.Context, db *pgx.Conn,
	fixture notificationCutoverFixture,
) (int64, string, string) {
	t.Helper()
	contract, err := NewCutoverContract()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `SET SESSION AUTHORIZATION harpy_adopter`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(ctx, `RESET SESSION AUTHORIZATION`) }()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `SET LOCAL search_path TO notification_cutover,pg_catalog`)
	if err != nil {
		t.Fatal(err)
	}
	var applied int64
	var keySHA, stateSHA string
	err = tx.QueryRow(ctx, contract.Query(), fixture.operationID, fixture.runID,
		"tenant", fixture.tenantID, int16(1), "notification", "notification_snapshot",
		"0/10", "0/20", "0/20", "0/20", "42", 1, "complete",
		[]byte(`{"id":"32000000-0000-4000-8000-000000000001"}`),
		fixture.batch, fixture.batchSHA, int64(len(fixture.batch)),
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Scan(
		&applied, &keySHA, &stateSHA)
	if err != nil {
		t.Fatalf("execute signed notification composite: %v", err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return applied, keySHA, stateSHA
}

func assertNotificationCutoverProjection(t *testing.T, ctx context.Context,
	db *pgx.Conn, fixture notificationCutoverFixture, wantSHA string,
) {
	t.Helper()
	var messages, effects, deliveries, attempts, receipts int
	var writer, lifecycle, targetSHA, bridge string
	err := db.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM notification_cutover.notifications_messages),
 (SELECT count(*) FROM notification_cutover.notifications_effect_receipts),
 (SELECT count(*) FROM notification_cutover.notifications_deliveries),
 (SELECT count(*) FROM notification_cutover.notifications_delivery_attempts),
 (SELECT count(*) FROM notification_cutover.notifications_v1_cutover_receipts),
 authority.writer,authority.lifecycle_state,receipt.target_root_sha256,receipt.bridge_state
FROM notification_cutover.notifications_writer_authority AS authority
JOIN notification_cutover.notifications_v1_cutover_receipts AS receipt ON true
WHERE authority.singleton AND receipt.operation_id=$1 AND receipt.run_id=$2`,
		fixture.operationID, fixture.runID).Scan(&messages, &effects, &deliveries,
		&attempts, &receipts, &writer, &lifecycle, &targetSHA, &bridge)
	if err != nil || messages != 1 || effects != 1 || deliveries != 1 || attempts != 1 ||
		receipts != 1 || writer != "v2" || lifecycle != "writable" || targetSHA != wantSHA ||
		bridge != "disabled_missing_signed_mtls_contract" {
		t.Fatalf("projection=%d/%d/%d/%d/%d %s/%s sha=%s bridge=%s err=%v",
			messages, effects, deliveries, attempts, receipts, writer, lifecycle,
			targetSHA, bridge, err)
	}
}

func assertNotificationCutoverTamperDenied(t *testing.T, ctx context.Context,
	db *pgx.Conn, fixture notificationCutoverFixture,
) {
	t.Helper()
	tampered := fixture
	tampered.batch = append(append([]byte(nil), fixture.batch...), byte(' '))
	_, err := db.Exec(ctx, `SET SESSION AUTHORIZATION harpy_adopter`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = db.Exec(ctx, `RESET SESSION AUTHORIZATION`) }()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, _ = tx.Exec(ctx, `SET LOCAL search_path TO notification_cutover,pg_catalog`)
	contract, _ := NewCutoverContract()
	_, err = tx.Exec(ctx, contract.Query(), tampered.operationID, tampered.runID,
		"tenant", tampered.tenantID, int16(1), "notification", "notification_snapshot",
		"0/10", "0/20", "0/20", "0/20", "42", 1, "complete",
		[]byte(`{"id":"32000000-0000-4000-8000-000000000001"}`),
		tampered.batch, fixture.batchSHA, int64(len(tampered.batch)),
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assertNotificationPGCode(t, err, "22023")
}

func assertNotificationLegacySurfacesDenied(t *testing.T, ctx context.Context,
	db *pgx.Conn, tenantID uuid.UUID,
) {
	t.Helper()
	for _, test := range []struct {
		role, query string
		args        []any
	}{
		{"harpy_adopter", `UPDATE notifications_messages SET subject='tampered'`, nil},
		{"harpy_adopter", `CREATE TABLE forbidden_notification_ddl(id integer)`, nil},
		{"harpy_adopter", `SELECT notifications_v1_adoption_provider_evidence_v1($1)`, []any{tenantID}},
		{"harpy_runtime", `INSERT INTO notifications_migration_intents(
producer,effect_key_hash,recipient_id,producer_generation,notification_generation,
canonical_digest,payload_digest,payload,job_id,state,created_at,updated_at,row_version)
VALUES('legacy',decode(repeat('0',64),'hex'),$1,1,1,decode(repeat('0',64),'hex'),
decode(repeat('0',64),'hex'),'{}',gen_random_uuid(),'queued',now(),now(),1)`, []any{tenantID}},
	} {
		_, err := db.Exec(ctx, `SET SESSION AUTHORIZATION `+test.role)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = tx.Exec(ctx, `SET LOCAL search_path TO notification_cutover,pg_catalog`)
		_, err = tx.Exec(ctx, test.query, test.args...)
		assertNotificationPGCode(t, err, "42501")
		_ = tx.Rollback(ctx)
		if _, err = db.Exec(ctx, `RESET SESSION AUTHORIZATION`); err != nil {
			t.Fatal(err)
		}
	}
}

func assertNotificationCutoverReceiptImmutable(t *testing.T, ctx context.Context,
	db *pgx.Conn, fixture notificationCutoverFixture,
) {
	t.Helper()
	_, err := db.Exec(ctx, `UPDATE notification_cutover.notifications_v1_cutover_receipts
SET target_count=target_count+1 WHERE operation_id=$1 AND run_id=$2`,
		fixture.operationID, fixture.runID)
	assertNotificationPGCode(t, err, "55000")
}

func assertNotificationPGCode(t *testing.T, err error, code string) {
	t.Helper()
	var postgres *pgconn.PgError
	if !errors.As(err, &postgres) || postgres.Code != code {
		t.Fatalf("postgres error=%v want code=%s", err, code)
	}
}

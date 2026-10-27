//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertEncryptedSMTPTestIntent(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema, operationID string,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, schema, false)
	var gotID, state string
	var ciphertext []byte
	err := tx.QueryRow(ctx, `SELECT operation_id::text,lifecycle_state,target_ciphertext
 FROM notifications_smtp_test_intents WHERE idempotency_key=$1`,
		"tenant-test-email-0001").Scan(&gotID, &state, &ciphertext)
	_ = tx.Rollback(ctx)
	targetAbsent := !bytes.Contains(ciphertext, []byte("admin@example.test"))
	if err != nil || gotID != operationID || state != "queued" || !targetAbsent {
		t.Fatalf("test intent id=%s state=%s target_absent=%t err=%v",
			gotID, state, targetAbsent, err)
	}
}

func testSMTPSettingsDirectTableDenial(t *testing.T, ctx context.Context,
	connection *pgx.Conn, track smtpSettingsTrack,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, track.schema, false)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_adopter"); err != nil {
		t.Fatal(err)
	}
	var count int64
	err := tx.QueryRow(ctx,
		"SELECT count(*) FROM notifications_smtp_settings_receipts").Scan(&count)
	_ = tx.Rollback(ctx)
	if pgCode(err) != "42501" {
		t.Fatalf("adopter direct-table denial code=%s err=%v", pgCode(err), err)
	}
}

func assertSMTPSettingsTenantIsolation(t *testing.T, ctx context.Context,
	connection *pgx.Conn, tracks []smtpSettingsTrack,
) {
	t.Helper()
	if len(tracks) != 2 {
		t.Fatalf("tracks=%d", len(tracks))
	}
	digests := make([][]byte, 0, len(tracks))
	for _, track := range tracks {
		tx := beginSMTPSettingsTenant(t, ctx, connection, track.schema, false)
		var digest []byte
		var count int64
		err := tx.QueryRow(ctx, `SELECT request_digest,
 (SELECT count(*) FROM notifications_smtp_settings_receipts)
 FROM notifications_smtp_settings_receipts WHERE idempotency_key=$1`,
			smtpSettingsIdempotencyKey).Scan(&digest, &count)
		_ = tx.Rollback(ctx)
		if err != nil || count != 1 || len(digest) != 32 {
			t.Fatalf("%s receipt count=%d digest=%x err=%v",
				track.schema, count, digest, err)
		}
		digests = append(digests, digest)
	}
	if bytes.Equal(digests[0], digests[1]) {
		t.Fatal("tenant-scoped same idempotency key shared one receipt digest")
	}
}

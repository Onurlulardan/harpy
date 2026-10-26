//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const smtpSettingsIdempotencyKey = "tenant-settings-0001"

func TestNotificationsPG18SMTPSettingsFreshAndAdopted(t *testing.T) {
	ctx := context.Background()
	connection := startStoragePG18(t, ctx)
	tracks := prepareSMTPSettingsDatabase(t, ctx, connection)
	for _, track := range tracks {
		track := track
		t.Run(track.origin, func(t *testing.T) {
			testSMTPSettingsCrashRetry(t, ctx, connection, track)
			testSMTPSettingsConflictAndEncryptedIntent(t, ctx, connection, track)
			testSMTPSettingsDirectTableDenial(t, ctx, connection, track)
		})
	}
	assertSMTPSettingsTenantIsolation(t, ctx, connection, tracks)
}

func testSMTPSettingsCrashRetry(t *testing.T, ctx context.Context,
	connection *pgx.Conn, track smtpSettingsTrack,
) {
	t.Helper()
	digest := sha256.Sum256([]byte("smtp settings " + track.origin))
	applySMTPSettingsAttempt(t, ctx, connection, track.schema, digest, true)
	assertSMTPSettingsMutationCount(t, ctx, connection, track.schema, 1, 0, 0)
	applySMTPSettingsAttempt(t, ctx, connection, track.schema, digest, false)
	assertSMTPSettingsMutationCount(t, ctx, connection, track.schema, 2, 1, 1)

	tx := beginSMTPSettingsTenant(t, ctx, connection, track.schema, true)
	store := smtpSettingsStore(tx, uuid.MustParse(smtpSettingsTenantID))
	if err := store.LockSMTPSettings(ctx); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	receipt, found, err := store.FindSMTPSettingsReceipt(ctx,
		smtpSettingsIdempotencyKey)
	if err != nil || !found || receipt.RequestDigest != digest ||
		receipt.ResultSettingsRevision != 8 || receipt.ProviderRevision != 2 {
		t.Fatalf("replay receipt=%+v found=%t err=%v", receipt, found, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertSMTPSettingsMutationCount(t, ctx, connection, track.schema, 2, 1, 1)
}

func applySMTPSettingsAttempt(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string,
	digest [32]byte, rollback bool,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, schema, true)
	store := smtpSettingsStore(tx, uuid.MustParse(smtpSettingsTenantID))
	if err := store.LockSMTPSettings(ctx); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	mutateSMTPSettingsProvider(t, ctx, tx)
	configID := smtpDomainID(t, smtpSettingsConfigID)
	actorID := smtpDomainID(t, smtpSettingsActorID)
	if err := store.AppendAudit(ctx, "notifications.provider.smtp_settings.replace",
		configID, actorID, 2); err != nil {
		t.Fatal(err)
	}
	receipt := ports.SMTPSettingsReceipt{IdempotencyKey: smtpSettingsIdempotencyKey,
		RequestDigest: digest, ExpectedSettingsRevision: 7, ResultSettingsRevision: 8,
		ProviderConfigID: configID, ProviderRevision: 2}
	if err := store.InsertSMTPSettingsReceipt(ctx, receipt); err != nil {
		t.Fatal(err)
	}
	if rollback {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func assertSMTPSettingsMutationCount(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string,
	revision, receipts, audits int64,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, schema, false)
	var gotRevision, gotReceipts, gotAudits, versions int64
	err := tx.QueryRow(ctx, `SELECT row_version,
 (SELECT count(*) FROM notifications_smtp_settings_receipts),
 (SELECT count(*) FROM notifications_audit_entries
  WHERE action='notifications.provider.smtp_settings.replace'),
 (SELECT count(*) FROM notifications_provider_config_versions WHERE config_id=$1)
 FROM notifications_provider_configs WHERE id=$1`, smtpSettingsConfigID).
		Scan(&gotRevision, &gotReceipts, &gotAudits, &versions)
	_ = tx.Rollback(ctx)
	if err != nil || gotRevision != revision || gotReceipts != receipts ||
		gotAudits != audits || versions != revision {
		t.Fatalf("revision=%d receipts=%d audits=%d versions=%d err=%v",
			gotRevision, gotReceipts, gotAudits, versions, err)
	}
}

func testSMTPSettingsConflictAndEncryptedIntent(t *testing.T, ctx context.Context,
	connection *pgx.Conn, track smtpSettingsTrack,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, track.schema, true)
	store := smtpSettingsStore(tx, uuid.MustParse(smtpSettingsTenantID))
	if err := store.LockSMTPSettings(ctx); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	existing, found, err := store.FindSMTPSettingsReceipt(ctx,
		smtpSettingsIdempotencyKey)
	different := sha256.Sum256([]byte("different request"))
	if err != nil || !found || existing.RequestDigest == different {
		t.Fatalf("same-key conflict precondition receipt=%+v found=%t err=%v",
			existing, found, err)
	}
	conflict := existing
	conflict.RequestDigest = different
	if err = store.InsertSMTPSettingsReceipt(ctx, conflict); pgCode(err) != "23505" {
		t.Fatalf("same-key different-digest code=%s err=%v", pgCode(err), err)
	}
	_ = tx.Rollback(ctx)

	tx = beginSMTPSettingsTenant(t, ctx, connection, track.schema, true)
	store = smtpSettingsStore(tx, uuid.MustParse(smtpSettingsTenantID))
	operationID := smtpDomainID(t, smtpTestOperationID)
	intent := ports.SMTPTestIntent{OperationID: operationID,
		ProviderVersionID: smtpDomainID(t, smtpSettingsVersion2),
		ActorID:           smtpDomainID(t, smtpSettingsActorID), IdempotencyKey: "tenant-test-email-0001",
		RequestDigest: sha256.Sum256([]byte("test intent")), ExpectedSettingsRevision: 8,
		EncryptedTarget: ports.EncryptedSecret{ID: operationID,
			Ciphertext: bytes.Repeat([]byte{0xa5}, 32), Nonce: bytes.Repeat([]byte{0xb6}, 12),
			WrappedDEK: bytes.Repeat([]byte{0xc7}, 32), KEKRevision: 1,
			Digest: sha256.Sum256([]byte("admin@example.test"))},
		RequestedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	if err = store.InsertSMTPTestIntent(ctx, intent); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertEncryptedSMTPTestIntent(t, ctx, connection, track.schema, operationID.String())
}

func pgCode(err error) string {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code
	}
	return ""
}

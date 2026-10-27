//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func testStorageRetention(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	now := time.Date(2028, 1, 2, 4, 0, 0, 0, time.UTC)
	tenant := uuid.MustParse("019b0000-0000-7000-8000-000000000040")
	kept := "019b0000-0000-7000-8000-000000000041"
	old := "019b0000-0000-7000-8000-000000000042"
	delivery := "019b0000-0000-7000-8000-000000000043"
	tx := beginStorageTenant(t, ctx, connection)
	seedRetentionRows(t, ctx, tx, now, tenant, kept, old, delivery)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	claim := retentionClaim(t, tenant)
	store, err := NewRetentionStore(RetentionDatabaseAccess{
		QueryRow: func(ctx context.Context, sql string, args ...any) RetentionRow {
			return tx.QueryRow(ctx, sql, args...)
		},
		RowNotFound: func(err error) bool { return errors.Is(err, pgx.ErrNoRows) },
	})
	if err != nil {
		t.Fatal(err)
	}
	cycle := retentionapp.Cycle{Claim: claim, CompletedAt: now,
		SuccessorJobID: uuid.MustParse("019b0000-0000-7000-8000-000000000049")}
	outcome, err := store.ApplyCycle(ctx, cycle)
	if err != nil || outcome.Counts.ContentRedacted < 1 ||
		outcome.Counts.AttemptsPurged < 1 || outcome.Counts.ReceiptsPurged < 1 ||
		outcome.Counts.EffectsPurged < 1 || outcome.Counts.DeliveriesPurged < 1 ||
		outcome.Counts.MessagesPurged < 1 || outcome.Counts.HintsPurged < 1 ||
		outcome.Counts.PlatformRecordsPurged < 3 {
		t.Fatalf("retention outcome=%+v err=%v", outcome, err)
	}
	replayCycle := cycle
	replayCycle.SuccessorJobID = uuid.MustParse("019b0000-0000-7000-8000-000000000050")
	replay, err := store.ApplyCycle(ctx, replayCycle)
	if err != nil || replay.Successor.JobID != outcome.Successor.JobID ||
		replay.CheckpointRevision != outcome.CheckpointRevision {
		t.Fatalf("retention replay=%+v err=%v", replay, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertRetentionRows(t, ctx, connection, kept, old, tenant, claim.JobID)
}

func retentionClaim(t *testing.T, tenant uuid.UUID) retentionapp.Claim {
	t.Helper()
	reference, digest, err := retentionapp.Reference(tenant)
	if err != nil {
		t.Fatal(err)
	}
	return retentionapp.Claim{
		JobID:     uuid.MustParse("019b0000-0000-7000-8000-000000000047"),
		AttemptID: uuid.MustParse("019b0000-0000-7000-8000-000000000048"),
		TenantID:  tenant, Kind: retentionapp.JobKind,
		PayloadRevision:  retentionapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}
}

func seedRetentionRows(t *testing.T, ctx context.Context, tx pgx.Tx, now time.Time,
	tenant uuid.UUID, kept, old, delivery string,
) {
	t.Helper()
	recipient := "019b0000-0000-7000-8000-000000000044"
	seedStorageDelivery(t, ctx, tx, now.Add(-400*24*time.Hour), old, delivery, recipient)
	_, err := tx.Exec(ctx, `UPDATE notifications_messages SET subject='',body='',
 purged_at=$2,row_version=2 WHERE id=$1`, old, now.Add(-366*24*time.Hour))
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_messages (
 id,recipient_id,template_version_id,producer,notification_type,category,
 template_kind,locale,subject,body,data_digest,subject_digest,body_digest,
 row_version,inbox_revision,created_at,dismissed_at
) VALUES ($1,$2,'12000000-0000-4000-8000-000000000002','workflow','task.kept',
 'workflow_action','plain_text','tr','kept','kept',decode(repeat('41',32),'hex'),
 decode(repeat('42',32),'hex'),decode(repeat('43',32),'hex'),1,2,$3,$4)`,
			kept, recipient, now.Add(-40*24*time.Hour), now.Add(-31*24*time.Hour))
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_effect_receipts
 (producer,effect_key_hash,recipient_id,canonical_digest,message_id,outcome,
  created_at,retain_until) VALUES
 ('workflow',decode(repeat('51',32),'hex'),$1,decode(repeat('52',32),'hex'),$2,
  'created',$3,$4),
 ('workflow',decode(repeat('53',32),'hex'),$1,decode(repeat('54',32),'hex'),$5,
  'created',$6,$7)`, recipient, kept, now.Add(-40*24*time.Hour), now.Add(400*24*time.Hour),
			old, now.Add(-400*24*time.Hour), now.Add(-35*24*time.Hour))
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_delivery_attempts
 (id,delivery_id,attempt_number,lease_generation,provider_kind,response_class,
 provider_reference_digest,status_code,error_code,started_at,finished_at)
 VALUES ('019b0000-0000-7000-8000-000000000045',$1,1,1,'smtp','accepted',
 decode(repeat('55',32),'hex'),'250','SMTP_ACCEPTED',$2,$3)`, delivery,
			now.Add(-100*24*time.Hour-time.Minute), now.Add(-100*24*time.Hour))
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_delivery_receipts
 (id,delivery_id,provider_reference_digest,outcome,signature_key_id,received_at)
 VALUES ('019b0000-0000-7000-8000-000000000046',$1,
 decode(repeat('55',32),'hex'),'delivered','receipt-key',$2)`, delivery,
			now.Add(-100*24*time.Hour))
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_realtime_hints
 (id,message_id,recipient_id,inbox_revision,created_at,published_at)
 VALUES ('019b0000-0000-7000-8000-000000000051',$1,$2,1,$3,$3)`,
			old, recipient, now.Add(-366*24*time.Hour))
	}
	if err == nil {
		err = seedRetentionPlatform(ctx, tx, now, tenant)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func seedRetentionPlatform(ctx context.Context, tx pgx.Tx, now time.Time,
	tenant uuid.UUID,
) error {
	terminal := now.Add(-25 * time.Hour)
	_, err := tx.Exec(ctx, `INSERT INTO platform.notification_realtime_tickets
 (id,tenant_id,user_id,token_hash,authorization_version,origin,created_at,expires_at,consumed_at)
 VALUES ('019b0000-0000-7000-8000-000000000052',$1,
 '019b0000-0000-7000-8000-000000000053',decode(repeat('61',32),'hex'),7,
 'https://app.example.com',$2,$3,$4)`, tenant, terminal.Add(-30*time.Second),
		terminal.Add(30*time.Second), terminal)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO platform.notification_provider_locators
 (provider_kind,provider_account_digest,external_message_digest,tenant_id,
 delivery_id,accepted_at,expires_at) VALUES ('smtp',decode(repeat('62',32),'hex'),
 decode(repeat('63',32),'hex'),$1,'019b0000-0000-7000-8000-000000000043',$2,$3)`,
			tenant, now.Add(-91*24*time.Hour), now.Add(-24*time.Hour))
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO platform.notification_legacy_connection_hits
 (id,tenant_id,user_id,consumer_manifest_digest,occurred_at) VALUES
 ('019b0000-0000-7000-8000-000000000054',$1,
 '019b0000-0000-7000-8000-000000000053',decode(repeat('64',32),'hex'),$2)`,
			tenant, now.Add(-91*24*time.Hour))
	}
	return err
}

func assertRetentionRows(t *testing.T, ctx context.Context, connection *pgx.Conn,
	kept, old string, tenant, jobID uuid.UUID,
) {
	t.Helper()
	tx := beginStorageTenant(t, ctx, connection)
	defer tx.Rollback(ctx)
	var subject, body string
	var keptMessages, keptEffects, oldMessages, tickets, audits, locators, hits, receipts int64
	err := tx.QueryRow(ctx, `SELECT subject,body,
 (SELECT count(*) FROM notifications_messages WHERE id=$1),
 (SELECT count(*) FROM notifications_effect_receipts WHERE message_id=$1),
 (SELECT count(*) FROM notifications_messages WHERE id=$2),
 (SELECT count(*) FROM platform.notification_realtime_tickets WHERE tenant_id=$3),
 (SELECT count(*) FROM platform.notification_realtime_ticket_audits WHERE tenant_id=$3),
 (SELECT count(*) FROM platform.notification_provider_locators WHERE tenant_id=$3),
 (SELECT count(*) FROM platform.notification_legacy_connection_hits WHERE tenant_id=$3),
 (SELECT count(*) FROM notifications_retention_receipts WHERE job_id=$4)
 FROM notifications_messages WHERE id=$1`, kept, old, tenant, jobID).Scan(
		&subject, &body, &keptMessages, &keptEffects, &oldMessages, &tickets,
		&audits, &locators, &hits, &receipts)
	if err != nil || subject != "" || body != "" || keptMessages != 1 ||
		keptEffects != 1 || oldMessages != 0 || tickets != 0 || audits != 1 ||
		locators != 0 || hits != 0 || receipts != 1 {
		t.Fatalf("retention rows kept=%d/%d old=%d ticket=%d/%d locator=%d hit=%d receipt=%d err=%v",
			keptMessages, keptEffects, oldMessages, tickets, audits, locators, hits,
			receipts, err)
	}
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `UPDATE notifications_retention_receipts
 SET has_remaining=NOT has_remaining WHERE job_id=$1`, jobID); err == nil {
		t.Fatal("runtime mutated immutable retention receipt")
	}
}

//go:build integration

package v1adoption

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type notificationCutoverFixture struct {
	tenantID, operationID, runID uuid.UUID
	batch                        []byte
	batchSHA                     string
}

func prepareNotificationCutoverFixture(t *testing.T, ctx context.Context,
	db *pgx.Conn,
) notificationCutoverFixture {
	t.Helper()
	fixture := notificationCutoverFixture{
		tenantID:    uuid.MustParse("31000000-0000-4000-8000-000000000001"),
		operationID: uuid.MustParse("31000000-0000-4000-8000-000000000002"),
		runID:       uuid.MustParse("31000000-0000-4000-8000-000000000003"),
	}
	prepareNotificationTenant(t, ctx, db, "notification_cutover", true)
	_, err := db.Exec(ctx, `GRANT USAGE ON SCHEMA notification_cutover TO harpy_runtime;
CREATE TABLE notification_cutover.tenancy_schema_identity(
 singleton boolean,tenant_id uuid,schema_name text,origin text)`)
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO notification_cutover.tenancy_schema_identity
VALUES(true,$1,'notification_cutover','v1_adopted')`, fixture.tenantID)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.cutover_operations VALUES(
$1,$2,'tenant',$3,'notifications','notifications',2,
repeat('a',64),'copying','2026-07-21T00:00:00Z')`,
			fixture.operationID, fixture.runID, fixture.tenantID)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.cutover_operation_checkpoint_keys
VALUES($1,$2,1,'notification')`, fixture.operationID, fixture.runID)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.cutover_copy_checkpoints
VALUES($1,$2,1,'notification','initial_copy',NULL)`, fixture.operationID, fixture.runID)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.async_v1_adoption_runs VALUES(
'31000000-0000-4000-8000-000000000004','tenant',$1,'notifications','notifications',
2,repeat('b',64),repeat('c',64),0,1,1,1,'accepted')`, fixture.tenantID)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.async_v1_zero_inventory_receipts VALUES(
'31000000-0000-4000-8000-000000000004',repeat('d',64),repeat('b',64),0,0,0,0)`)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.async_v1_restart_fence_receipts VALUES
('31000000-0000-4000-8000-000000000004','enqueue','notifications-api',repeat('e',64)),
('31000000-0000-4000-8000-000000000004','scheduler','notifications-scheduler',repeat('f',64)),
('31000000-0000-4000-8000-000000000004','consumer','notifications-worker',repeat('1',64))`)
	}
	if err == nil {
		_, err = db.Exec(ctx, `INSERT INTO platform.contract_retirements VALUES(
'notifications_actor_stream_v1','notifications','notifications',0,
'2026-07-20T00:00:00Z','eligible',0,1)`)
	}
	if err != nil {
		t.Fatalf("seed notification cutover fixture: %v", err)
	}
	fixture.batch = notificationCutoverBatch(t, ctx, db, fixture)
	digest := sha256.Sum256(fixture.batch)
	fixture.batchSHA = hex.EncodeToString(digest[:])
	return fixture
}

func notificationCutoverBatch(t *testing.T, ctx context.Context, db *pgx.Conn,
	fixture notificationCutoverFixture,
) []byte {
	t.Helper()
	messageID := "32000000-0000-4000-8000-000000000001"
	deliveryID := "32000000-0000-4000-8000-000000000002"
	attemptID := "32000000-0000-4000-8000-000000000003"
	digest := `\x` + string(make([]byte, 0)) + hex.EncodeToString(make([]byte, 32))
	created := "2026-07-20T01:02:03Z"
	message := normalizeNotificationRecord(t, ctx, db, "notifications_messages", map[string]any{
		"id": messageID, "recipient_id": "32000000-0000-4000-8000-000000000004",
		"template_version_id": "12000000-0000-4000-8000-000000000002",
		"producer":            "legacy_v1", "notification_type": "workflow_notification",
		"category": "workflow_action", "template_kind": "plain_text", "locale": "tr",
		"subject": "Task ready", "body": "Review the task", "data_digest": digest,
		"subject_digest": digest, "body_digest": digest, "row_version": 1,
		"inbox_revision": 1, "created_at": created, "read_at": nil,
		"dismissed_at": nil, "purged_at": nil,
	})
	effect := normalizeNotificationRecord(t, ctx, db, "notifications_effect_receipts", map[string]any{
		"producer": "legacy_v1", "effect_key_hash": digest,
		"recipient_id":     "32000000-0000-4000-8000-000000000004",
		"canonical_digest": digest, "message_id": messageID,
		"legacy_receipt_reference": "v1:notifications:" + messageID,
		"outcome":                  "created", "created_at": created, "retain_until": "2027-07-22T01:02:03Z",
	})
	delivery := normalizeNotificationRecord(t, ctx, db, "notifications_deliveries", map[string]any{
		"id": deliveryID, "message_id": messageID, "provider_config_version_id": nil,
		"channel": "email", "state": "accepted", "attempt": 1, "maximum_attempts": 8,
		"lease_owner": "", "lease_generation": 1, "lease_until": nil,
		"not_before": created, "outcome_code": "V1_ACCEPTED", "outcome_unknown": false,
		"mandatory": false, "row_version": 2,
	})
	attempt := normalizeNotificationRecord(t, ctx, db, "notifications_delivery_attempts", map[string]any{
		"id": attemptID, "delivery_id": deliveryID, "attempt_number": 1,
		"lease_generation": 1, "provider_kind": "smtp",
		"response_class": "legacy_v1_accepted", "provider_reference_digest": digest,
		"status_code": "", "error_code": "V1_ACCEPTED",
		"started_at": created, "finished_at": created,
	})
	sourceSHA := sha256.Sum256([]byte("notification-v1-source"))
	changes := []any{
		notificationCutoverChange(1, "message", messageID, messageID, message, sourceSHA),
		notificationCutoverChange(2, "effect", messageID, messageID, effect, sourceSHA),
		notificationCutoverChange(3, "delivery", deliveryID, deliveryID, delivery, sourceSHA),
		notificationCutoverChange(4, "attempt", attemptID, attemptID, attempt, sourceSHA),
	}
	batch, err := json.Marshal(map[string]any{
		"operation_id": fixture.operationID, "run_id": fixture.runID,
		"scope_kind": "tenant", "scope_id": fixture.tenantID,
		"relation_ordinal": 1, "checkpoint_key": "notification",
		"checkpoint_state": "complete", "last_key": map[string]any{"id": messageID},
		"source_position": map[string]any{"snapshot_name": "notification_snapshot",
			"snapshot_lsn": "0/10", "received_lsn": "0/20", "commit_lsn": "0/20",
			"final_lsn": "0/20", "source_xid": "42", "batch_ordinal": 1},
		"changes": changes,
	})
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func normalizeNotificationRecord(t *testing.T, ctx context.Context, db *pgx.Conn,
	table string, record map[string]any,
) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var normalized []byte
	query := `SELECT to_jsonb(jsonb_populate_record(NULL::notification_cutover.` + table + `,$1::jsonb))`
	if err = db.QueryRow(ctx, query, raw).Scan(&normalized); err != nil {
		t.Fatalf("normalize %s: %v", table, err)
	}
	return normalized
}

func notificationCutoverChange(ordinal int, kind, identity, targetID string,
	record json.RawMessage, sourceSHA [32]byte,
) map[string]any {
	key := map[string]any{"source_identity": "v1-notification-1",
		"source_kind": "notification", "target_identity": identity, "target_kind": kind}
	value := map[string]any{"source_identity": "v1-notification-1",
		"source_kind": "notification", "source_sha256": hex.EncodeToString(sourceSHA[:]),
		"target_id": targetID, "target_identity": identity, "target_kind": kind,
		"target_record": record}
	return map[string]any{"ordinal": ordinal, "kind": "insert", "key": key,
		"before": nil, "after": value}
}

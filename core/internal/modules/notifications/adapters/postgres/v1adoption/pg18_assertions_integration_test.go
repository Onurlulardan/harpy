//go:build integration

package v1adoption

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func seedV1Notifications(t *testing.T, ctx context.Context, connection *pgx.Conn,
	schema string, tenantID, userID uuid.UUID,
) {
	t.Helper()
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	_, err := connection.Exec(ctx, `INSERT INTO `+schema+`.notifications
(id,tenant_id,user_id,type,level,title,message,data,channel,is_read,created_at,read_at)
VALUES
('20000000-0000-4000-8000-000000000001',$1,$2,'password_reset','info',
 'Security','Reset password','{"source":"v1"}','email',true,$3,$3),
('20000000-0000-4000-8000-000000000002',$1,$2,'workflow_notification','info',
 'Workflow','Task ready','{"source":"v1"}','flow',false,$4,NULL),
('20000000-0000-4000-8000-000000000003',$1,$2,'toast','info',
 'Toast','Ephemeral','{}','websocket',false,$5,NULL)`,
		tenantID, userID, created, created.Add(time.Minute), created.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
}

func assertNotificationCounts(t *testing.T, ctx context.Context,
	connection *pgx.Conn, messages, attempts int,
) {
	t.Helper()
	var gotMessages, gotAttempts int
	err := connection.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM notification_a.notifications_messages),
 (SELECT count(*) FROM notification_a.notifications_delivery_attempts)`).Scan(
		&gotMessages, &gotAttempts)
	if err != nil || gotMessages != messages || gotAttempts != attempts {
		t.Fatalf("messages=%d attempts=%d err=%v", gotMessages, gotAttempts, err)
	}
}

func assertAdoptedEvidence(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	var state string
	var adopted, issues int
	var equal bool
	err := connection.QueryRow(ctx, `SELECT run.lifecycle_state, checkpoint.adopted_count,
 checkpoint.source_checksum = checkpoint.target_checksum,
 (SELECT count(*) FROM notification_a.notifications_adoption_issues)
FROM notification_a.notifications_adoption_runs run
JOIN notification_a.notifications_adoption_checkpoints checkpoint ON checkpoint.run_id=run.id`).Scan(
		&state, &adopted, &equal, &issues)
	if err != nil || state != "verified" || adopted != 2 || !equal || issues != 0 {
		t.Fatalf("state=%s adopted=%d equal=%t issues=%d err=%v", state, adopted, equal, issues, err)
	}
	var stable uuid.UUID
	if err = connection.QueryRow(ctx, `SELECT id FROM notification_a.notifications_messages
WHERE id='20000000-0000-4000-8000-000000000001'`).Scan(&stable); err != nil {
		t.Fatalf("stable V1 ID missing: %v", err)
	}
}

func assertAttemptImmutableForAdopter(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	transaction, err := connection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	_, err = transaction.Exec(ctx, `SET LOCAL ROLE harpy_adopter;
SET LOCAL search_path TO notification_a, pg_catalog;
UPDATE notifications_delivery_attempts SET status_code='tampered';`)
	var postgres *pgconn.PgError
	if !errors.As(err, &postgres) || postgres.Code != "42501" {
		t.Fatalf("immutable attempt update err=%v", err)
	}
}

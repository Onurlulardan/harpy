//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	storageTenantSchema = "tenant_notifications_storage"
	storageMessageID    = "019b0000-0000-7000-8000-000000000001"
	storageDeliveryID   = "019b0000-0000-7000-8000-000000000002"
)

func prepareStorageDatabase(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	_, err := connection.Exec(ctx, `CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_adopter NOLOGIN;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE SCHEMA `+storageTenantSchema+` AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_runtime;
GRANT USAGE ON SCHEMA `+storageTenantSchema+` TO harpy_runtime, harpy_adopter;`)
	if err != nil {
		t.Fatal(err)
	}
	applyStorageMigration(t, ctx, connection, "platform", "")
	applyStorageMigration(t, ctx, connection, "tenant", storageTenantSchema)
}

func applyStorageMigration(t *testing.T, ctx context.Context, connection *pgx.Conn,
	scope, schema string,
) {
	t.Helper()
	directory := filepath.Join("..", "..", "migrations", scope, "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".fresh.sql") {
			continue
		}
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			t.Fatal(readErr)
		}
		tx, beginErr := connection.Begin(ctx)
		if beginErr == nil && schema != "" {
			_, beginErr = tx.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)", schema+", pg_catalog")
		}
		if beginErr == nil {
			_, beginErr = tx.Exec(ctx, string(document))
		}
		if beginErr != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply %s: %v", entry.Name(), beginErr)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func seedStorageDelivery(t *testing.T, ctx context.Context, tx pgx.Tx, now time.Time,
	messageID, deliveryID, recipientID string,
) {
	t.Helper()
	_, err := tx.Exec(ctx, `INSERT INTO notifications_messages (
 id, recipient_id, template_version_id, producer, notification_type, category,
 template_kind, locale, subject, body, data_digest, subject_digest, body_digest,
 row_version, inbox_revision, created_at
) VALUES ($1,$2,$3,'workflow','task.ready','workflow_action','plain_text','tr',
 'ready','ready',decode(repeat('11',32),'hex'),decode(repeat('22',32),'hex'),
 decode(repeat('33',32),'hex'),1,1,$4);`, messageID, recipientID,
		"12000000-0000-4000-8000-000000000002", now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO notifications_deliveries (
 id, message_id, provider_config_version_id, channel, state, attempt,
 maximum_attempts, lease_owner, lease_generation, lease_until, not_before,
 outcome_code, outcome_unknown, mandatory, row_version
) VALUES ($1,$2,NULL,'email','pending',0,8,'',0,NULL,$3,'',false,true,1);`,
		deliveryID, messageID, now)
	if err != nil {
		t.Fatal(err)
	}
}

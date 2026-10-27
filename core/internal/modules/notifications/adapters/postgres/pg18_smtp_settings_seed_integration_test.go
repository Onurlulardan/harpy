//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	smtpSettingsTenantID = "019d0000-0000-7000-8000-000000000001"
	smtpSettingsConfigID = "019d0000-0000-7000-8000-000000000002"
	smtpSettingsSecretID = "019d0000-0000-7000-8000-000000000003"
	smtpSettingsVersion1 = "019d0000-0000-7000-8000-000000000004"
	smtpSettingsVersion2 = "019d0000-0000-7000-8000-000000000005"
	smtpSettingsActorID  = "019d0000-0000-7000-8000-000000000006"
	smtpTestOperationID  = "019d0000-0000-7000-8000-000000000007"
)

func seedSMTPSettingsProvider(t *testing.T, ctx context.Context,
	connection *pgx.Conn, schema string,
) {
	t.Helper()
	tx := beginSMTPSettingsTenant(t, ctx, connection, schema, false)
	_, err := tx.Exec(ctx, `INSERT INTO notifications_provider_secrets (
 id,ciphertext,nonce,wrapped_dek,kek_revision,plaintext_digest,created_at
) VALUES ($1,decode(repeat('11',17),'hex'),decode(repeat('22',12),'hex'),
 decode(repeat('33',28),'hex'),1,decode(repeat('44',32),'hex'),
 pg_catalog.transaction_timestamp());`, smtpSettingsSecretID)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_provider_configs (
 id,provider_kind,name,enabled,current_version_id,row_version,created_at,updated_at
 ) VALUES ($1,'smtp','Tenant SMTP',true,NULL,1,
 pg_catalog.transaction_timestamp(),pg_catalog.transaction_timestamp());`,
			smtpSettingsConfigID)
	}
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notifications_provider_config_versions (
 id,config_id,version_number,secret_id,metadata,metadata_digest,created_at
) VALUES ($3,$2,1,$1,
 '{"smtp_host":"smtp.example.test","smtp_username":"mailer","smtp_security":"starttls","from_address":"noreply@example.test","smtp_port":587}'::jsonb,
	 decode(repeat('55',32),'hex'),pg_catalog.transaction_timestamp());`,
			smtpSettingsSecretID, smtpSettingsConfigID, smtpSettingsVersion1)
	}
	if err == nil {
		_, err = tx.Exec(ctx,
			`UPDATE notifications_provider_configs SET current_version_id=$1 WHERE id=$2`,
			smtpSettingsVersion1, smtpSettingsConfigID)
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func mutateSMTPSettingsProvider(t *testing.T, ctx context.Context, tx pgx.Tx) {
	t.Helper()
	result, err := tx.Exec(ctx, `WITH version AS (
 INSERT INTO notifications_provider_config_versions (
  id,config_id,version_number,secret_id,metadata,metadata_digest,created_at
 ) VALUES ($1,$2,2,$3,
  '{"smtp_host":"mail.example.test","smtp_username":"mailer","smtp_security":"starttls","from_address":"noreply@example.test","smtp_port":587}'::jsonb,
  decode(repeat('66',32),'hex'),pg_catalog.transaction_timestamp())
 RETURNING id
)
UPDATE notifications_provider_configs SET current_version_id=version.id,
 row_version=2,updated_at=pg_catalog.transaction_timestamp()
FROM version WHERE notifications_provider_configs.id=$2 AND row_version=1;`,
		smtpSettingsVersion2, smtpSettingsConfigID, smtpSettingsSecretID)
	if err != nil || result.RowsAffected() != 1 {
		t.Fatalf("mutate provider affected=%d err=%v", result.RowsAffected(), err)
	}
}

//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const retentionTenantBID = "72000000-0000-4000-8000-000000000082"

type retentionFixture struct {
	connection             *sql.Conn
	tenantA, tenantB       uuid.UUID
	authorityA, authorityB uuid.UUID
	dueA, futureA, dueB    uuid.UUID
}

type retentionOutcome struct {
	successor                                   uuid.UUID
	runAt                                       time.Time
	reference, digest, dedupe                   string
	priority                                    int
	redacted, raw, entitlement, financial, held int64
	checkpoint                                  int64
}

func prepareRetentionFixture(
	t *testing.T,
	ctx context.Context,
) retentionFixture {
	t.Helper()
	database := prepareBillingOrigin(t, ctx, "fresh")
	if _, err := database.ExecContext(ctx,
		"GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE",
	); err != nil {
		t.Fatal(err)
	}
	connection := billingRuntimeConnection(t, ctx, database)
	now := time.Now().UTC().Truncate(time.Microsecond)
	fixture := retentionFixture{
		connection: connection, tenantA: uuid.MustParse(billingTenantID),
		tenantB:    uuid.MustParse(retentionTenantBID),
		authorityA: billingV7(t), authorityB: billingV7(t),
		dueA: billingV7(t), futureA: billingV7(t), dueB: billingV7(t),
	}
	if _, err := connection.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,$3)
`, fixture.authorityA, fixture.tenantA, now); err != nil {
		t.Fatal(err)
	}
	seedRetentionTenantB(t, ctx, database, fixture.tenantB)
	if _, err := connection.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,$3)
`, fixture.authorityB, fixture.tenantB, now); err != nil {
		t.Fatal(err)
	}
	providerID := billingV7(t)
	if _, err := connection.ExecContext(ctx, `
INSERT INTO platform.billing_provider_accounts(
 id,authority_id,provider_key,provider_account_key,lifecycle,
 supports_checkout,supports_webhooks,supports_proration,supports_refunds,
 created_at,disabled_at,revision
) VALUES($1,$2,'retention','retention-test','active',false,true,false,false,
 $3,NULL,1)`, providerID, billingGlobalAuthorityID(t, ctx, database), now); err != nil {
		t.Fatal(err)
	}
	seedRetentionReceipt(t, ctx, connection, fixture.authorityA, fixture.tenantA,
		providerID, fixture.dueA, "a-due", now.Add(-time.Minute), now)
	seedRetentionReceipt(t, ctx, connection, fixture.authorityA, fixture.tenantA,
		providerID, fixture.futureA, "a-future", now.Add(time.Hour), now)
	seedRetentionReceipt(t, ctx, connection, fixture.authorityB, fixture.tenantB,
		providerID, fixture.dueB, "b-due", now.Add(-time.Minute), now)
	return fixture
}

func seedRetentionTenantB(
	t *testing.T, ctx context.Context, database *sql.DB, tenantID uuid.UUID,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,'t_retention_b','ACTIVE',NULL)`, tenantID); err != nil {
		t.Fatal(err)
	}
	fixture := newBillingCutoverAuthorityFixture(
		"tenant", tenantID, "billing_forms", 1,
		"v2", "v2", billingAcceptanceA, "billing",
	)
	if err = seedBillingCutoverAuthority(ctx, transaction, fixture); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func seedRetentionReceipt(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID, tenantID, providerID, receiptID uuid.UUID,
	eventKey string, bodyExpiry, now time.Time,
) {
	t.Helper()
	mappingID := billingV7(t)
	receivedAt := now.AddDate(0, 0, -31)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_provider_objects(
 id,authority_id,tenant_id,provider_account_id,object_kind,provider_object_id,
 subscription_id,invoice_id,created_at,retired_at,revision
) VALUES($1,$2,$3,$4,'customer',$5,NULL,NULL,$6,NULL,1)`,
		mappingID, authorityID, tenantID, providerID, eventKey, receivedAt)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_webhook_receipts(
 id,authority_id,tenant_id,provider_account_id,provider_object_mapping_id,
 provider_event_id,body_sha256,authentication_digest,verification_revision,
 event_occurred_at,signature_issued_at,algorithm,nonce,encrypted_body,wrapped_dek,
 key_id,verified_at,received_at,processor_state,processed_at,raw_body_expires_at,
 retain_until,revision
) VALUES($7,$2,$3,$4,$1,$5,$8,$8,1,$6,$6,'AES-256-GCM',$9,$10,$11,
 'retention-key',$6,$6,'applied',$6,$12,$13,1)`,
			mappingID, authorityID, tenantID, providerID, eventKey, receivedAt,
			receiptID, billingDigest(eventKey), make([]byte, 12), make([]byte, 17),
			[]byte{1}, bodyExpiry, now.AddDate(1, 0, 0))
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func executeRetention(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID, tenantID, jobID, attemptID, successorID uuid.UUID,
	reference, digest string,
) retentionOutcome {
	t.Helper()
	var result retentionOutcome
	err := connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_execute_retention_cycle_v1($1,$2,$3,$4,$5,$6,$7)
`, authorityID, tenantID, jobID, attemptID, successorID, reference, digest).Scan(
		&result.successor, &result.runAt, &result.reference, &result.digest,
		&result.dedupe, &result.priority, &result.redacted, &result.raw,
		&result.entitlement, &result.financial, &result.held, &result.checkpoint,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func retentionDigest(character string) string {
	return strings.Repeat(character, 64)
}

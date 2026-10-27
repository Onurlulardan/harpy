//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingProviderIngressSQLAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixtureConnection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	tenant := prepareBillingTenantFixture(t, ctx, fixtureConnection, globalAuthorityID)
	fixture := prepareBillingProviderFixture(
		t, ctx, fixtureConnection, globalAuthorityID, tenant,
	)
	assertProviderTargetResolved(t, ctx, fixtureConnection, fixture)
	assertConcurrentProviderReceipt(t, ctx, database, fixture)
	assertStaleProviderTargetRejected(t, ctx, fixtureConnection, fixture)
}

func assertProviderTargetResolved(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture billingProviderFixture,
) {
	t.Helper()
	statement := loadBillingAdapterSQL(t, "provider", "resolve_provider_target.sql")
	var tenantID, mappingID, accountID, subscriptionID uuid.UUID
	var invoiceID *uuid.UUID
	var mappingRevision, groupGeneration, billingGeneration int64
	err := connection.QueryRowContext(
		ctx, statement, "testpay", fixture.accountKey,
		"subscription", fixture.objectID,
	).Scan(
		&tenantID, &mappingID, &accountID, &subscriptionID, &invoiceID,
		&mappingRevision, &groupGeneration, &billingGeneration,
	)
	if err != nil || tenantID != uuid.MustParse(billingTenantID) ||
		mappingID != fixture.mappingID || accountID != fixture.providerAccountID ||
		subscriptionID != fixture.subscriptionID || invoiceID != nil ||
		mappingRevision != 1 || groupGeneration != 1 ||
		billingGeneration != 1 {
		t.Fatalf("resolved provider target is invalid: %v", err)
	}
}

func assertConcurrentProviderReceipt(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingProviderFixture,
) {
	t.Helper()
	statement := loadBillingAdapterSQL(t, "provider", "insert_webhook_receipt.sql")
	connections := []*sql.Conn{
		billingRuntimeConnection(t, ctx, database),
		billingRuntimeConnection(t, ctx, database),
	}
	bodies := [][]byte{billingDigest("provider-body-a"), billingDigest("provider-body-b")}
	arguments := [][]any{
		providerReceiptArguments(t, fixture, bodies[0]),
		providerReceiptArguments(t, fixture, bodies[1]),
	}
	start := make(chan struct{})
	results := make(chan providerInsertResult, 2)
	var group sync.WaitGroup
	for index := range connections {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			result, err := connections[index].ExecContext(
				ctx, statement, arguments[index]...,
			)
			if err != nil {
				results <- providerInsertResult{err: err}
				return
			}
			rows, err := result.RowsAffected()
			results <- providerInsertResult{rows: rows, err: err}
		}(index)
	}
	close(start)
	group.Wait()
	close(results)
	var inserted int64
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		inserted += result.rows
	}
	if inserted != 1 {
		t.Fatalf("concurrent provider receipt inserts = %d, want 1", inserted)
	}
	var count int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_webhook_receipts
WHERE provider_account_id=$1 AND provider_event_id='evt_provider_fixture'
`, fixture.providerAccountID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("durable provider receipt count=%d error=%v", count, err)
	}
}

type providerInsertResult struct {
	rows int64
	err  error
}

func providerReceiptArguments(
	t *testing.T,
	fixture billingProviderFixture,
	bodySHA []byte,
) []any {
	t.Helper()
	receivedAt := fixture.now.Add(time.Minute)
	return []any{
		billingV7(t), fixture.authorityID, uuid.MustParse(billingTenantID),
		fixture.providerAccountID, fixture.mappingID, "evt_provider_fixture",
		bodySHA, billingDigest("provider-auth"), 1,
		receivedAt.Add(-time.Second), receivedAt.Add(-2 * time.Second),
		"AES-256-GCM", []byte("123456789012"), make([]byte, 17),
		[]byte("wrapped-provider-dek"), "billing-kek-v1", receivedAt, receivedAt,
		receivedAt.Add(30 * 24 * time.Hour), receivedAt.Add(400 * 24 * time.Hour),
	}
}

func assertStaleProviderTargetRejected(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture billingProviderFixture,
) {
	t.Helper()
	_, err := connection.ExecContext(ctx, `
UPDATE platform.domain_cutovers SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='forms'
`, uuid.MustParse(billingTenantID))
	if err != nil {
		t.Fatal(err)
	}
	statement := loadBillingAdapterSQL(t, "provider", "resolve_provider_target.sql")
	err = connection.QueryRowContext(
		ctx, statement, "testpay", fixture.accountKey,
		"subscription", fixture.objectID,
	).Scan(new(uuid.UUID), new(uuid.UUID), new(uuid.UUID), new(uuid.UUID),
		new(*uuid.UUID), new(int64), new(int64), new(int64), new(int64))
	if err != sql.ErrNoRows {
		t.Fatalf("stale provider target error=%v, want no rows", err)
	}
}

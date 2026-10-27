//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type invoiceRaceResult struct {
	replay bool
	err    error
}

type invoiceStatements struct {
	lockDraft         string
	findFinalization  string
	writeFinalization string
}

type lockedInvoice struct {
	revision  int64
	finalized bool
	policy    []byte
}

func TestBillingInvoiceFinalizationRaceAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixtureConnection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	tenant := prepareBillingTenantFixture(t, ctx, fixtureConnection, globalAuthorityID)
	subscription := prepareBillingSubscriptionFixture(
		t, ctx, fixtureConnection, globalAuthorityID, tenant,
	)
	fixture := prepareBillingInvoiceFixture(t, ctx, fixtureConnection, subscription)
	statements := invoiceStatements{
		lockDraft:         loadBillingAdapterSQL(t, "invoice", "lock_draft.sql"),
		findFinalization:  loadBillingAdapterSQL(t, "invoice", "find_finalization.sql"),
		writeFinalization: loadBillingAdapterSQL(t, "invoice", "write_finalization.sql"),
	}
	sourceOperationID := billingV7(t)
	start := make(chan struct{})
	results := make(chan invoiceRaceResult, 2)
	for range 2 {
		connection := billingRuntimeConnection(t, ctx, database)
		go func() {
			<-start
			results <- finalizeInvoice(
				ctx, connection, fixture, sourceOperationID, statements,
			)
		}()
	}
	close(start)
	first, second := <-results, <-results
	for _, result := range []invoiceRaceResult{first, second} {
		if result.err != nil {
			t.Fatal(result.err)
		}
	}
	if first.replay == second.replay {
		t.Fatalf("invoice replay flags=%v,%v, want one write and one replay",
			first.replay, second.replay)
	}
	assertInvoiceFinalization(t, ctx, database, fixture, sourceOperationID)
}

func finalizeInvoice(
	ctx context.Context,
	connection *sql.Conn,
	fixture billingInvoiceFixture,
	sourceOperationID uuid.UUID,
	statements invoiceStatements,
) invoiceRaceResult {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return invoiceRaceResult{err: err}
	}
	fail := func(err error) invoiceRaceResult {
		_ = transaction.Rollback()
		return invoiceRaceResult{err: err}
	}
	locked, err := lockInvoiceDraft(ctx, transaction, fixture, statements.lockDraft)
	if err != nil {
		return fail(fmt.Errorf("lock invoice draft: %w", err))
	}
	found, err := findInvoiceFinalization(
		ctx, transaction, fixture, sourceOperationID, statements.findFinalization,
	)
	if err != nil {
		return fail(err)
	}
	if found {
		return invoiceRaceResult{replay: true, err: transaction.Commit()}
	}
	if locked.finalized || locked.revision != 1 {
		return fail(fmt.Errorf("invoice finalized without operation receipt"))
	}
	var returnedID uuid.UUID
	err = transaction.QueryRowContext(ctx, statements.writeFinalization,
		fixture.authorityID, uuid.MustParse(billingTenantID), fixture.invoiceID,
		int64(1), int64(2), int64(1050), int64(0), int64(1050), int64(0),
		int64(1), locked.policy, fixture.now.Add(time.Minute), sourceOperationID,
	).Scan(&returnedID)
	if err != nil || returnedID != fixture.invoiceID {
		return fail(fmt.Errorf("write invoice finalization id=%s: %w", returnedID, err))
	}
	return invoiceRaceResult{err: transaction.Commit()}
}

func lockInvoiceDraft(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingInvoiceFixture,
	statement string,
) (lockedInvoice, error) {
	result := lockedInvoice{}
	var invoiceID uuid.UUID
	var currency string
	var registryVersion, settled int64
	var linesDocument []byte
	err := transaction.QueryRowContext(ctx, statement, fixture.authorityID,
		uuid.MustParse(billingTenantID), fixture.invoiceID,
	).Scan(&invoiceID, &result.revision, &currency, &registryVersion, &settled,
		&result.policy, &result.finalized, &linesDocument)
	if err == nil && len(linesDocument) <= 2 {
		err = fmt.Errorf("invoice lines document is empty")
	}
	return result, err
}

func findInvoiceFinalization(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingInvoiceFixture,
	sourceOperationID uuid.UUID,
	statement string,
) (bool, error) {
	var invoiceID, sourceID uuid.UUID
	var expectedRevision, resultRevision int64
	var invoiceTotal, settled, amountDue, creditBalance, registryVersion int64
	var currency string
	var policy []byte
	var finalizedAt time.Time
	err := transaction.QueryRowContext(ctx, statement, fixture.authorityID,
		uuid.MustParse(billingTenantID), sourceOperationID,
	).Scan(&invoiceID, &sourceID, &expectedRevision, &resultRevision,
		&invoiceTotal, &settled, &amountDue, &creditBalance, &currency,
		&registryVersion, &policy, &finalizedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func assertInvoiceFinalization(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingInvoiceFixture,
	sourceOperationID uuid.UUID,
) {
	t.Helper()
	var state string
	var revision, total, due, finalizations int64
	err := database.QueryRowContext(ctx, `
SELECT invoice.state,invoice.revision,invoice.invoice_total_minor_units,
       invoice.amount_due_minor_units,
       (SELECT count(*) FROM platform.billing_invoice_finalizations
        WHERE source_operation_id=$2)
FROM platform.billing_invoices AS invoice WHERE invoice.id=$1
`, fixture.invoiceID, sourceOperationID).Scan(
		&state, &revision, &total, &due, &finalizations,
	)
	if err != nil || state != "finalized" || revision != 2 || total != 1050 ||
		due != 1050 || finalizations != 1 {
		t.Fatalf("invoice state=%s revision=%d total=%d due=%d receipts=%d err=%v",
			state, revision, total, due, finalizations, err)
	}
}

package invoice

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	"github.com/google/uuid"
)

func TestFinalizeLocksCalculatesAndWritesInOneBoundStore(t *testing.T) {
	authorityID, invoiceID, lineID, operationID := testV7(), testV7(), testV7(), testV7()
	now := time.Date(2026, 7, 17, 7, 0, 0, 0, time.UTC)
	digest := sha256.Sum256([]byte("invoice-policy"))
	storage := testStore(
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{
			invoiceID, int64(3), "USD", int64(1), int64(500), digest[:], false,
			[]byte(`[{"id":"` + lineID.String() +
				`","gross":1000,"discount":100,"tax":180,` +
				`"debit":0,"credit":0,"total":1080}]`),
		}},
		scriptedRow{values: []any{authorityID}},
		scriptedRow{err: noRows{}},
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{invoiceID}},
	)
	storage.notFound = func(err error) bool { _, ok := err.(noRows); return ok }
	finalizer, err := invoiceapp.NewFinalizer(
		storage, uuid.NewV7, func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := finalizer.Finalize(context.Background(), invoiceapp.FinalizeRequest{
		InvoiceID: invoiceID, ActorID: testV7(),
		ExpectedRevision: 3, SourceOperationID: operationID,
	})
	if err != nil || result.Replay ||
		result.Finalization.Totals.InvoiceTotal().MinorUnits() != 1080 ||
		result.Finalization.Totals.AmountDue().MinorUnits() != 580 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

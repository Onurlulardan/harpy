package discount

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
)

func TestRedemptionInsertRequiresReturnedIdentity(t *testing.T) {
	authorityID, redemptionID := testV7(), testV7()
	storage, executions := testStore(
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{redemptionID}},
	)
	redemption := discountapp.Redemption{
		ID: redemptionID, DiscountVersionID: testV7(), SubscriptionID: testV7(),
		OfferingVersionID: testV7(), PriceVersionID: testV7(),
		SourceOperationID: testV7(), PolicySHA256: sha256.Sum256([]byte("policy")),
		RedeemedAt: time.Date(2026, 7, 17, 6, 0, 0, 0, time.UTC),
	}
	if err := storage.createRedemption(context.Background(), redemption); err != nil ||
		*executions != 1 {
		t.Fatalf("executions=%d err=%v", *executions, err)
	}
}

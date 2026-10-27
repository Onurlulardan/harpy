package discount

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
)

func TestRedeemLocksPlansAndCreatesInOneBoundStore(t *testing.T) {
	authorityID := testV7()
	discountVersionID, subscriptionID := testV7(), testV7()
	offeringID, priceID, operationID := testV7(), testV7(), testV7()
	redemptionID := testV7()
	now := time.Date(2026, 7, 17, 6, 0, 0, 0, time.UTC)
	digest := sha256.Sum256([]byte("policy"))
	storage, _ := testStore(
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{
			discountVersionID, "LAUNCH", offeringID, priceID, "USD",
			now.Add(-time.Hour), now.Add(time.Hour), (*int64)(nil), int64(0),
			(*int64)(nil), int64(0), true, digest[:],
		}},
		scriptedRow{values: []any{authorityID}},
		scriptedRow{err: errNoRowsForTest{}},
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{redemptionID}},
	)
	storage.notFound = func(err error) bool { _, ok := err.(errNoRowsForTest); return ok }
	code, _ := discountdomain.NormalizeCode("launch")
	request := discountapp.RedeemRequest{Code: code, SubscriptionID: subscriptionID,
		OfferingVersionID: offeringID, PriceVersionID: priceID,
		Currency: "USD", SourceOperationID: operationID}
	result, err := storage.Redeem(context.Background(),
		discountapp.RedeemExecution{Request: request, Now: now},
		func(snapshot discountapp.EligibilitySnapshot, existing *discountapp.Redemption) (
			discountapp.RedemptionPlan, error,
		) {
			if existing != nil || snapshot.PolicySHA256 != digest {
				t.Fatal("unexpected redemption planning state")
			}
			value := discountapp.Redemption{ID: redemptionID,
				DiscountVersionID: discountVersionID, SubscriptionID: subscriptionID,
				OfferingVersionID: offeringID, PriceVersionID: priceID,
				SourceOperationID: operationID, PolicySHA256: digest, RedeemedAt: now}
			return discountapp.RedemptionPlan{
				Result: discountapp.RedeemResult{Redemption: value}, Create: &value,
			}, nil
		})
	if err != nil || result.Redemption.ID != redemptionID {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

type errNoRowsForTest struct{}

func (errNoRowsForTest) Error() string { return "no rows" }

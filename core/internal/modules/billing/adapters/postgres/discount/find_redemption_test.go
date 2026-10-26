package discount

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"
)

func TestFindRedemptionMapsImmutableSnapshot(t *testing.T) {
	authorityID := testV7()
	id, versionID, subscriptionID := testV7(), testV7(), testV7()
	offeringID, priceID, operationID := testV7(), testV7(), testV7()
	digest := sha256.Sum256([]byte("policy"))
	now := time.Date(2026, 7, 17, 6, 0, 0, 0, time.UTC)
	storage, _ := testStore(
		scriptedRow{values: []any{authorityID}},
		scriptedRow{values: []any{
			id, versionID, subscriptionID, offeringID, priceID,
			operationID, digest[:], now,
		}},
	)
	result, found, err := storage.findRedemption(context.Background(), operationID)
	if err != nil || !found || result.ID != id || result.PolicySHA256 != digest {
		t.Fatalf("result=%+v found=%v err=%v", result, found, err)
	}
}

//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertRedemptionRaceState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingSubscriptionFixture,
	discountVersionID uuid.UUID,
	sourceOperationID uuid.UUID,
	findStatement string,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_discount_redemptions
WHERE discount_version_id=$1
`, discountVersionID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("discount redemption count=%d err=%v", count, err)
	}
	connection := billingRuntimeConnection(t, ctx, database)
	var redemptionID, versionID, subscriptionID, offeringID, priceID, sourceID uuid.UUID
	var policy []byte
	var redeemedAt time.Time
	err := connection.QueryRowContext(ctx, findStatement, fixture.authorityID,
		uuid.MustParse(billingTenantID), sourceOperationID,
	).Scan(&redemptionID, &versionID, &subscriptionID, &offeringID,
		&priceID, &sourceID, &policy, &redeemedAt)
	if err != nil || versionID != discountVersionID || sourceID != sourceOperationID {
		t.Fatalf("find discount redemption version=%s source=%s err=%v", versionID, sourceID, err)
	}
}

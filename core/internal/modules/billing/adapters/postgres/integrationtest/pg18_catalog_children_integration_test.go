//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertCatalogChildren(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	discountVersionID uuid.UUID,
	campaignVersionID uuid.UUID,
) {
	t.Helper()
	var discountOfferings int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_discount_offerings
WHERE discount_version_id=$1
`, discountVersionID).Scan(&discountOfferings); err != nil {
		t.Fatal(err)
	}
	var campaignDiscounts int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_campaign_discounts
WHERE campaign_version_id=$1
`, campaignVersionID).Scan(&campaignDiscounts); err != nil {
		t.Fatal(err)
	}
	if discountOfferings != 1 || campaignDiscounts != 1 {
		t.Fatalf("catalog children discount=%d campaign=%d",
			discountOfferings, campaignDiscounts)
	}
}

func assertCampaignChildCountRollback(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	authorityID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := callCreateCampaign(ctx, transaction, authorityID, actorID,
		billingV7Must(), billingV7Must(), "campaign-missing-child", now)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("create child-count probe: %v", err)
	}
	err = transaction.Commit()
	if billingPGCode(err) != "23514" {
		t.Fatalf("missing campaign child commit code=%q result=%+v err=%v",
			billingPGCode(err), result, err)
	}
}

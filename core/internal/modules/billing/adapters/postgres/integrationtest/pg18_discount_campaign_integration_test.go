//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type catalogResult struct {
	aggregateID uuid.UUID
	versionID   uuid.UUID
	revision    int64
	version     int64
	lifecycle   string
	replay      bool
}

func TestBillingDiscountCampaignRoutinesAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	connection := billingRuntimeConnection(t, ctx, database)
	authorityID := billingGlobalAuthorityID(t, ctx, database)
	actorID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	offering := createOffering(t, ctx, connection, authorityID, actorID,
		"catalog-offering", billingDigest("catalog-offering"), now)
	discount := createDiscount(t, ctx, connection, authorityID, actorID,
		offering.versionID, now)
	campaign := createCampaign(t, ctx, connection, authorityID, actorID,
		discount.versionID, "campaign-create", now)
	if campaign.revision != 1 || campaign.version != 1 || campaign.replay {
		t.Fatalf("unexpected campaign result: %+v", campaign)
	}
	replay, err := callCreateCampaign(ctx, connection, authorityID, actorID,
		billingV7Must(), billingV7Must(), "campaign-create", now)
	if err != nil || !replay.replay || replay.aggregateID != campaign.aggregateID {
		t.Fatalf("unexpected campaign replay: result=%+v err=%v", replay, err)
	}
	assertCatalogChildren(t, ctx, database, discount.versionID, campaign.versionID)
	assertCampaignChildCountRollback(t, ctx, connection, authorityID, actorID, now)
}

func createDiscount(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID uuid.UUID, actorID uuid.UUID, offeringVersionID uuid.UUID,
	now time.Time,
) catalogResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	result := catalogResult{}
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_discount_v2(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
 $14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26
)`, authorityID, billingV7Must(), billingV7Must(), billingV7Must(), actorID,
		"discount-create", billingDigest("discount-create"), "INTEGRATION10",
		"Integration discount", "Integration discount fixture", true,
		"basis_point", nil, nil, nil, 1000, "once", nil, 10, nil, nil,
		false, int16(1), now, nil, now,
	).Scan(&result.aggregateID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_discount_offering_v1($1,$2,$3,$4)
`, authorityID, result.versionID, offeringVersionID, now)
	}
	finishCatalogTransaction(t, transaction, err)
	return result
}

func createCampaign(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID uuid.UUID, actorID uuid.UUID, discountVersionID uuid.UUID,
	key string, now time.Time,
) catalogResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := callCreateCampaign(ctx, transaction, authorityID, actorID,
		billingV7Must(), billingV7Must(), key, now)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_campaign_discount_v1($1,$2,$3,$4)
`, authorityID, result.versionID, discountVersionID, now)
	}
	finishCatalogTransaction(t, transaction, err)
	return result
}

func callCreateCampaign(
	ctx context.Context, connection billingQueryRower,
	authorityID uuid.UUID, actorID uuid.UUID, campaignID uuid.UUID,
	versionID uuid.UUID, key string, now time.Time,
) (catalogResult, error) {
	result := catalogResult{}
	err := connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_campaign_v2(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,
 $13::jsonb,pg_catalog.sha256(pg_catalog.convert_to($13::jsonb::text,'UTF8')),
 $14,$15,$16,$17,$18,$19
)`, authorityID, billingV7Must(), campaignID, versionID, actorID, key,
		billingDigest(key), "campaign-"+campaignID.String(), 10, int16(1), now, nil,
		`{"name":"Integration campaign","description":"Integration campaign fixture","banner_text":"Save now","auto_apply":true}`,
		"Integration campaign",
		"Integration campaign fixture", "Save now", true, true, now,
	).Scan(&result.aggregateID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	return result, err
}

func finishCatalogTransaction(t *testing.T, transaction *sql.Tx, err error) {
	t.Helper()
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

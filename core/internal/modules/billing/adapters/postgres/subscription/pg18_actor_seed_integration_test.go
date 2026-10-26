//go:build integration

package subscription

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type subscriptionActorFixture struct {
	tenantID, authorityID, actorID uuid.UUID
	subscriptionID                 uuid.UUID
	now, periodEnd                 time.Time
}

func seedSubscriptionActor(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) subscriptionActorFixture {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin actor seed transaction: %v", err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL ROLE harpy_owner`); err != nil {
		t.Fatalf("assume actor seed owner role: %v", err)
	}
	now := time.Date(2026, 7, 17, 1, 10, 0, 0, time.UTC)
	fixture := subscriptionActorFixture{
		tenantID:    uuid.MustParse("71000000-0000-4000-8000-000000000071"),
		authorityID: uuid.Must(uuid.NewV7()), actorID: uuid.New(),
		subscriptionID: uuid.Must(uuid.NewV7()), now: now,
		periodEnd: now.AddDate(0, 1, 0),
	}
	var globalAuthorityID uuid.UUID
	if err = transaction.QueryRowContext(ctx, `
SELECT authority_id FROM platform.billing_global_authorities
WHERE scope_kind='global' AND authority_group_id='billing'
`).Scan(&globalAuthorityID); err != nil {
		t.Fatalf("load seeded global billing authority: %v", err)
	}
	if _, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,1,$3)
`, fixture.authorityID, fixture.tenantID, now.Add(-time.Hour)); err != nil {
		t.Fatalf("register tenant subscription authority: %v", err)
	}
	offeringID, offeringVersionID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	priceVersionID := uuid.Must(uuid.NewV7())
	operationID := uuid.Must(uuid.NewV7())
	digest := sha256.Sum256([]byte("subscription-actor-offering"))
	var gotOffering, gotVersion uuid.UUID
	var revision, version int64
	var lifecycle string
	var replay bool
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)`, globalAuthorityID, operationID, offeringID, offeringVersionID,
		fixture.actorID, "subscription-actor", digest[:], "subscription-actor",
		"Subscription Actor", false, true, 0, 0, int16(1), int16(0),
		now.Add(-time.Hour), nil, now.Add(-time.Hour),
	).Scan(&gotOffering, &gotVersion, &revision, &version, &lifecycle, &replay)
	if err != nil || gotOffering != offeringID || gotVersion != offeringVersionID {
		t.Fatalf("create actor offering = %v/%v, %v", gotOffering, gotVersion, err)
	}
	if _, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_price_v1($1,$2,$3,'monthly',1000,'USD',1,$4)
`, globalAuthorityID, offeringVersionID, priceVersionID, now.Add(-time.Hour)); err != nil {
		t.Fatalf("create actor price: %v", err)
	}
	providerID := uuid.Must(uuid.NewV7())
	_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_provider_accounts(
 id,authority_id,provider_key,provider_account_key,lifecycle,
 supports_checkout,supports_webhooks,supports_proration,supports_refunds,
 created_at,disabled_at,revision
) VALUES($1,$2,'manual',$3,'active',false,false,false,false,$4,NULL,1)
`, providerID, globalAuthorityID, "actor-"+providerID.String(), now.Add(-time.Hour))
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscriptions(
 id,authority_id,tenant_id,provider_account_id,offering_version_id,
 price_version_id,state,period_start,period_end,trial_ends_at,grace_ends_at,
 cancel_at_period_end,cancel_at,created_at,updated_at,revision
) VALUES($1,$2,$3,$4,$5,$6,'active',$7,$8,NULL,NULL,false,NULL,$7,$7,1)
`, fixture.subscriptionID, fixture.authorityID, fixture.tenantID, providerID,
			offeringVersionID, priceVersionID, now.Add(-time.Hour), fixture.periodEnd)
	}
	if err != nil {
		t.Fatalf("seed actor subscription: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit actor subscription seed: %v", err)
	}
	return fixture
}

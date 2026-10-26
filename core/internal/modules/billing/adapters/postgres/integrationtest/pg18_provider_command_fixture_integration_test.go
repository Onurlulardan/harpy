//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type providerPGFixture struct {
	connection                        *sql.Conn
	tenantID, authorityID             uuid.UUID
	providerID, mappingID             uuid.UUID
	subscriptionID                    uuid.UUID
	offeringVersionID, priceVersionID uuid.UUID
	commandID, finalizerID            uuid.UUID
	actorID                           uuid.UUID
	now                               time.Time
	digest                            []byte
	digestHex                         string
}

func prepareProviderPGFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) providerPGFixture {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime; SET ROLE harpy_owner",
	); err != nil {
		t.Fatal(err)
	}
	fixture := providerPGFixture{
		connection: connection, tenantID: uuid.MustParse(providerPGTenant),
		authorityID: billingPGV7(t), providerID: billingPGV7(t),
		mappingID: billingPGV7(t), subscriptionID: billingPGV7(t),
		offeringVersionID: billingPGV7(t), priceVersionID: billingPGV7(t),
		commandID: billingPGV7(t), finalizerID: billingPGV7(t), actorID: uuid.New(),
		now: time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
	}
	digest := sha256.Sum256([]byte("provider-command-pg18"))
	fixture.digest = digest[:]
	fixture.digestHex = fmt.Sprintf("%x", digest)
	seedProviderPGBilling(t, ctx, fixture)
	return fixture
}

func seedProviderPGBilling(t *testing.T, ctx context.Context, f providerPGFixture) {
	t.Helper()
	var globalAuthority uuid.UUID
	if err := f.connection.QueryRowContext(ctx, `
SELECT authority_id FROM platform.billing_global_authorities
`).Scan(&globalAuthority); err != nil {
		t.Fatal(err)
	}
	offeringID := billingPGV7(t)
	transaction, err := f.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1($1,$2,1,1,1,$3)
`, f.authorityID, f.tenantID, f.now); err == nil {
		_, err = transaction.ExecContext(ctx, `
 SELECT platform.billing_create_offering_v1(
 $1,$2,$3,$4,$5,'provider-fixture',$6,'provider-fixture','Provider Fixture',
 false,true,0,0,1::smallint,0::smallint,$7,NULL,$7
)`, globalAuthority, billingPGV7(t), offeringID, f.offeringVersionID,
			uuid.New(), sha256Bytes("offering"), f.now)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_price_v1($1,$2,$3,'monthly',1000,'USD',1,$4)
`, globalAuthority, f.offeringVersionID, f.priceVersionID, f.now)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_provider_accounts(
 id,authority_id,provider_key,provider_account_key,lifecycle,
 supports_checkout,supports_webhooks,supports_proration,supports_refunds,
 created_at,disabled_at,revision
) VALUES($1,$2,'testpay','acct_provider_pg18','active',false,true,false,false,$3,NULL,1)
`, f.providerID, globalAuthority, f.now)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_subscriptions(
 id,authority_id,tenant_id,provider_account_id,offering_version_id,
 price_version_id,state,period_start,period_end,cancel_at_period_end,
 created_at,updated_at,revision
) VALUES($1,$2,$3,$4,$5,$6,'active',$7,$8,false,$7,$7,1)
`, f.subscriptionID, f.authorityID, f.tenantID, f.providerID,
			f.offeringVersionID, f.priceVersionID, f.now, f.now.AddDate(0, 1, 0))
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `
INSERT INTO platform.billing_provider_objects(
 id,authority_id,tenant_id,provider_account_id,object_kind,provider_object_id,
 subscription_id,created_at,revision
) VALUES($1,$2,$3,$4,'subscription','sub_provider_pg18',$5,$6,1)
`, f.mappingID, f.authorityID, f.tenantID, f.providerID, f.subscriptionID, f.now)
	}
	if err != nil {
		t.Fatalf("seed provider Billing fixture: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func billingPGV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func sha256Bytes(value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return digest[:]
}

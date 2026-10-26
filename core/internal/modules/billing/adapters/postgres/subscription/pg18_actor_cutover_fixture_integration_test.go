//go:build integration

package subscription

import (
	"context"
	"database/sql"
	"testing"
)

func prepareSubscriptionCutoverFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, subscriptionCutoverFixtureSQL); err != nil {
		t.Fatalf("prepare subscription cutover fixture: %v", err)
	}
}

const subscriptionCutoverFixtureSQL = `
CREATE TABLE platform.contract_retirements (
 contract_key text PRIMARY KEY,owner_module text NOT NULL,domain text NOT NULL,
 state text NOT NULL DEFAULT 'active',remaining_scope_count bigint NOT NULL
);
CREATE FUNCTION platform.contract_retirement_provision_v1(
 p_contract_key text,p_owner_module text,p_domain text,
 p_remaining_scope_count bigint
) RETURNS void LANGUAGE plpgsql AS $fixture$
BEGIN
 INSERT INTO platform.contract_retirements(
  contract_key,owner_module,domain,remaining_scope_count
 ) VALUES(
  p_contract_key,p_owner_module,p_domain,p_remaining_scope_count
 ) ON CONFLICT(contract_key) DO NOTHING;
 IF NOT EXISTS(
  SELECT 1 FROM platform.contract_retirements AS retirement
  WHERE retirement.contract_key=p_contract_key
    AND retirement.owner_module=p_owner_module
    AND retirement.domain=p_domain
    AND retirement.remaining_scope_count=p_remaining_scope_count
 ) THEN
  RAISE EXCEPTION 'contract retirement provisioning drift';
 END IF;
END;
$fixture$;`

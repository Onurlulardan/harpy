//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestTenancyActivationStepCapabilitiesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startTenantMigrationDatabase(t, ctx)
			prepareTransactionCapabilityDatabase(t, ctx, database, origin)
			plan := newTransactionCapabilityPlan(t, "steps-"+origin)
			insertTransactionCapabilityPlan(t, ctx, database, plan)
			assertActivationStepCapabilityACL(t, ctx, database)
			assertActivationStepCommitGuard(t, ctx, database, plan)
			claimAndConsumeActivationSteps(t, ctx, database, plan)
		})
	}
}

func assertActivationStepCapabilityACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, signature := range []string{
		"platform.tenancy_activation_step_assert_v1(uuid,text)",
		"platform.tenancy_activation_step_consume_v1(uuid,text)",
	} {
		var allowed bool
		err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed)
		if err != nil || allowed {
			t.Fatalf("internal step capability ACL %s=%t err=%v",
				signature, allowed, err)
		}
	}
	var tableAllowed bool
	err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_table_privilege(
 'harpy_runtime','platform.tenancy_activation_step_capabilities',
 'SELECT,INSERT,UPDATE,DELETE'
)`).Scan(&tableAllowed)
	if err != nil || tableAllowed {
		t.Fatalf("runtime step capability table ACL=%t err=%v", tableAllowed, err)
	}
	for _, signature := range []string{
		"platform.tenancy_onboarding_activate_v3(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v5" +
			"(uuid,uuid,uuid,text,text,uuid,text,text,text,text)",
	} {
		var allowed bool
		err = database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',$1,'EXECUTE'
)`, signature).Scan(&allowed)
		if err != nil || !allowed {
			t.Fatalf("runtime successor ACL %s=%t err=%v", signature, allowed, err)
		}
	}
}

func assertActivationStepCommitGuard(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	claimActivationStepCapabilities(t, ctx, transaction, plan)
	requireActivationSQLState(
		t, transaction.Commit(), "23514", "unconsumed activation step capability",
	)
}

func claimAndConsumeActivationSteps(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	revision := claimActivationStepCapabilities(t, ctx, transaction, plan)
	var replayed bool
	err := transaction.QueryRowContext(ctx, `
SELECT platform.tenancy_activation_initialize_v5(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
)`, plan.operationID, plan.environmentID, plan.tenantID, plan.name, plan.slug,
		plan.ownerID, plan.schema, plan.purpose, plan.manifest, plan.kind,
	).Scan(&replayed)
	if err != nil || replayed {
		t.Fatalf("activation v5 replayed=%t err=%v", replayed, err)
	}
	consumeActivationBaseOffering(t, ctx, transaction, plan)
	for _, command := range activationStepCommands {
		var consumed bool
		err = transaction.QueryRowContext(ctx,
			`SELECT platform.test_consume_step_v1($1,$2)`,
			plan.operationID, command).Scan(&consumed)
		if err != nil || !consumed {
			t.Fatalf("consume %s=%t err=%v", command, consumed, err)
		}
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit consumed activation steps: %v", err)
	}
	assertNoActivationCapabilities(t, ctx, database, plan)
	return revision
}

func claimActivationStepCapabilities(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	plan transactionCapabilityPlan,
) int64 {
	t.Helper()
	if _, err := transaction.ExecContext(ctx, `
SELECT pg_catalog.set_config(
 'application_name','harpy-v2:core-operator-initialize',true
)`); err != nil {
		t.Fatal(err)
	}
	var revision int64
	err := transaction.QueryRowContext(ctx, `
SELECT revision FROM platform.tenancy_onboarding_activate_v3($1,$2,$3)`,
		plan.operationID, plan.leaseToken, plan.revision).Scan(&revision)
	if err != nil {
		t.Fatalf("claim activation step capabilities: %v", err)
	}
	return revision
}

func consumeActivationBaseOffering(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	plan transactionCapabilityPlan,
) {
	t.Helper()
	var consumed bool
	err := transaction.QueryRowContext(ctx,
		`SELECT platform.test_consume_activation_v1($1,$2)`,
		plan.operationID, plan.manifest).Scan(&consumed)
	if err != nil || !consumed {
		t.Fatalf("consume base offering=%t err=%v", consumed, err)
	}
}

func assertNoActivationCapabilities(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan,
) {
	t.Helper()
	var base, steps, tenants, facts int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT pg_catalog.count(*) FROM platform.tenancy_activation_command_capabilities
 WHERE operation_id=$1),
(SELECT pg_catalog.count(*) FROM platform.tenancy_activation_step_capabilities
 WHERE operation_id=$1),
(SELECT pg_catalog.count(*) FROM platform.tenants WHERE id=$2),
(SELECT pg_catalog.count(*) FROM platform.tenancy_activation_facts
 WHERE operation_id=$1)`, plan.operationID, plan.tenantID).Scan(
		&base, &steps, &tenants, &facts,
	)
	if err != nil || base != 0 || steps != 0 || tenants != 1 || facts != 1 {
		t.Fatalf("base/steps/tenants/facts=%d/%d/%d/%d err=%v",
			base, steps, tenants, facts, err)
	}
}

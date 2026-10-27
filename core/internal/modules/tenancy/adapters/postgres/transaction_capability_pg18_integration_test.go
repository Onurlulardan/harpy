//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestTenancyTransactionCapabilitiesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startTenantMigrationDatabase(t, ctx)
			prepareTransactionCapabilityDatabaseThrough(
				t, ctx, database, origin, 20260729100036,
			)
			plan := newTransactionCapabilityPlan(t, origin)
			insertTransactionCapabilityPlan(t, ctx, database, plan)
			assertTransactionCapabilityACL(t, ctx, database)
			assertActivationCapabilityGuard(t, ctx, database, plan, false)
			assertActivationCapabilityGuard(t, ctx, database, plan, true)
			revision := claimAndConsumeActivationCapability(t, ctx, database, plan)
			completeTransactionCapabilityActivation(
				t, ctx, database, plan, revision,
			)
			assertAuthorityCapabilityFlow(t, ctx, database, plan)
		})
	}
}

func assertTransactionCapabilityACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, signature := range []string{
		"platform.tenancy_activation_command_consume_v1" +
			"(uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text," +
			"timestamptz,bigint,text)",
		"platform.tenancy_authority_reconcile_scope_assert_v1(text)",
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege('harpy_runtime',$1,'EXECUTE')`,
			signature).Scan(&allowed); err != nil {
			t.Fatal(err)
		}
		if allowed {
			t.Fatalf("runtime can execute internal capability routine %s", signature)
		}
	}
	for _, table := range []string{
		"platform.tenancy_activation_command_capabilities",
		"platform.tenancy_authority_reconcile_scopes",
	} {
		var allowed bool
		if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_table_privilege(
 'harpy_runtime',$1,'SELECT,INSERT,UPDATE,DELETE'
)`, table).Scan(&allowed); err != nil {
			t.Fatal(err)
		}
		if allowed {
			t.Fatalf("runtime can access capability table %s", table)
		}
	}
}

func assertActivationCapabilityGuard(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan, consumeWrong bool,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	revision := claimActivationCapability(t, ctx, transaction, plan)
	if revision <= plan.revision {
		t.Fatalf("activation capability revision = %d", revision)
	}
	if consumeWrong {
		var consumed bool
		err := transaction.QueryRowContext(ctx,
			`SELECT platform.test_consume_activation_v1($1,$2)`,
			plan.operationID, strings.Repeat("f", 64)).Scan(&consumed)
		if err != nil || consumed {
			t.Fatalf("wrong capability consumed=%t err=%v", consumed, err)
		}
	}
	requireActivationSQLState(
		t, transaction.Commit(), "23514", "unconsumed activation capability",
	)
}

func claimAndConsumeActivationCapability(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan,
) int64 {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	revision := claimActivationCapability(t, ctx, transaction, plan)
	var consumed bool
	err := transaction.QueryRowContext(ctx,
		`SELECT platform.test_consume_activation_v1($1,$2)`,
		plan.operationID, plan.manifest).Scan(&consumed)
	if err != nil || !consumed {
		t.Fatalf("valid capability consumed=%t err=%v", consumed, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("valid activation capability commit: %v", err)
	}
	var remaining int
	if err = database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM platform.tenancy_activation_command_capabilities
WHERE operation_id=$1`, plan.operationID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("consumed capability rows = %d", remaining)
	}
	return revision
}

func claimActivationCapability(
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
SELECT revision FROM platform.tenancy_onboarding_activate_v2($1,$2,$3)`,
		plan.operationID, plan.leaseToken, plan.revision).Scan(&revision)
	if err != nil {
		t.Fatalf("claim activation capability: %v", err)
	}
	return revision
}

func completeTransactionCapabilityActivation(
	t *testing.T, ctx context.Context, database *sql.DB,
	plan transactionCapabilityPlan, revision int64,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	execute := func(statement string, arguments ...any) {
		t.Helper()
		if _, execErr := transaction.ExecContext(
			ctx, statement, arguments...,
		); execErr != nil {
			t.Fatalf("insert completed activation facts: %v", execErr)
		}
	}
	execute(`INSERT INTO platform.tenants(
 id,name,slug,status,settings,owner_id,schema_name,created_at,updated_at
) VALUES($1,$2,$3,'ACTIVE','{}',$4,$5,$6,$6)`,
		plan.tenantID, plan.name, plan.slug, plan.ownerID, plan.schema, now)
	execute(`INSERT INTO platform.tenancy_tenant_bindings(
 environment_id,tenant_id,purpose,created_at
) VALUES($1,$2,'saas_control',$3)`,
		plan.environmentID, plan.tenantID, now)
	execute(`INSERT INTO platform.tenancy_activation_facts(
 operation_id,environment_id,tenant_id,owner_user_id,binding_purpose,
 manifest_sha256,activated_at
) VALUES($1,$2,$3,$4,'saas_control',$5,$6)`,
		plan.operationID, plan.environmentID, plan.tenantID, plan.ownerID,
		plan.manifest, now)
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit completed activation facts: %v", err)
	}
	completeTransactionCapabilityOperation(
		t, ctx, database, plan.operationID, revision,
	)
}

func requireActivationSQLState(t *testing.T, err error, want, operation string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s unexpectedly succeeded", operation)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != want {
		t.Fatalf("%s SQLSTATE want=%s err=%v", operation, want, err)
	}
}

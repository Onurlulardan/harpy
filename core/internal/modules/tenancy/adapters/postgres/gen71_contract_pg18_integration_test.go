//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGen71TenancyContractAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startTenantMigrationDatabase(t, ctx)
			resetTenantMigrationCatalog(t, ctx, database)
			runPlatformMigrations(t, ctx, database, origin)
			insertDeploymentPolicy(t, ctx, database, "saas")
			assertGen71TenancyACL(t, ctx, database)

			fixture := newOnboardingFixture(t, "saas")
			arguments := fixture.acceptArguments(fixture.manifestSHA)
			_, _, err := callGen71Accept(
				t, ctx, database, onboardingAcceptSQL, arguments,
			)
			requireActivationSQLState(
				t, err, "42501", "retired onboarding accept",
			)
			requireGen71OnboardingRows(t, ctx, database, fixture.operationID, 0, 0)

			const successor = `SELECT * FROM platform.tenancy_onboarding_accept_v2(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31)`
			for attempt := 0; attempt < 2; attempt++ {
				state, revision, callErr := callGen71Accept(
					t, ctx, database, successor, arguments,
				)
				if callErr != nil || state != "accepted" || revision != 1 {
					t.Fatalf("successor accept %d=%s/%d err=%v",
						attempt, state, revision, callErr)
				}
			}
			requireGen71OnboardingRows(t, ctx, database, fixture.operationID, 1, 9)
		})
	}
}

func assertGen71TenancyACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, signature := range []string{
		"platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)",
		"platform.tenancy_onboarding_activate_v1(uuid,text,bigint)",
		"platform.tenancy_onboarding_activate_v2(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text)",
		"platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint)",
		"platform.tenancy_activation_assert_v1(uuid,uuid,name)",
	} {
		requireGen71FunctionACL(t, ctx, database, signature, false)
	}
	for _, signature := range []string{
		"platform.tenancy_onboarding_accept_v2(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)",
		"platform.tenancy_onboarding_activate_v3(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v5(uuid,uuid,uuid,text,text,uuid,text,text,text,text)",
		"platform.tenancy_onboarding_complete_v2(uuid,uuid,text,text,bigint)",
		"platform.tenancy_activation_assert_v2(uuid,uuid,name)",
	} {
		requireGen71FunctionACL(t, ctx, database, signature, true)
	}
}

func requireGen71FunctionACL(
	t *testing.T, ctx context.Context, database *sql.DB,
	signature string, runtimeAllowed bool,
) {
	t.Helper()
	var runtime, login bool
	err := database.QueryRowContext(ctx, `SELECT
pg_catalog.has_function_privilege('harpy_runtime',$1,'EXECUTE'),
pg_catalog.has_function_privilege('harpy_v2_runtime',$1,'EXECUTE')`,
		signature).Scan(&runtime, &login)
	if err != nil || runtime != runtimeAllowed || login {
		t.Fatalf("tenancy ACL %s=%t/%t want=%t/false err=%v",
			signature, runtime, login, runtimeAllowed, err)
	}
}

func callGen71Accept(
	t *testing.T, ctx context.Context, database *sql.DB,
	query string, arguments []any,
) (string, int64, error) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	err := transaction.QueryRowContext(ctx, query, arguments...).Scan(
		&operationID, &state, &revision, &updatedAt,
	)
	if err != nil {
		return "", 0, err
	}
	return state, revision, transaction.Commit()
}

func requireGen71OnboardingRows(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID uuid.UUID, wantOperations, wantSteps int,
) {
	t.Helper()
	var operations, steps int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.tenancy_onboarding_operations
 WHERE operation_id=$1),
(SELECT count(*) FROM platform.tenancy_onboarding_steps
 WHERE operation_id=$1)`, operationID).Scan(&operations, &steps)
	if err != nil || operations != wantOperations || steps != wantSteps {
		t.Fatalf("onboarding rows=%d/%d want=%d/%d err=%v",
			operations, steps, wantOperations, wantSteps, err)
	}
}

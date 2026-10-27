//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestActivationScopeProfilesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startTenantMigrationDatabase(t, ctx)
			prepareActivationScopeDatabaseThrough(
				t, ctx, database, origin, 20260729100036,
			)
			environment := newActivationScopeFixture(t,
				"environment_initialize", "saas_control", "control-"+origin)
			customer := newActivationScopeFixture(t,
				"saas_customer_tenant_provision", "saas_customer", "customer-"+origin)
			insertActivationOperation(t, ctx, database, environment)
			insertActivationOperation(t, ctx, database, customer)

			assertWrongActivationProfile(t, ctx, database, environment,
				"harpy-v2:core-worker")
			assertActivationAndReplay(t, ctx, database, environment,
				"harpy-v2:core-operator-initialize")
			assertWrongActivationProfile(t, ctx, database, customer,
				"harpy-v2:core-operator-initialize")
			assertActivationAndReplay(t, ctx, database, customer,
				"harpy-v2:core-worker")
		})
	}
}

func assertWrongActivationProfile(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture activationScopeFixture, application string,
) {
	t.Helper()
	if _, err := callActivationV4(t, ctx, database, fixture, application); !activationErrorDenied(err) {
		t.Fatalf("%s accepted wrong profile %q: %v", fixture.kind, application, err)
	}
	var tenants, domains, authorities int
	err := database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.tenants WHERE id=$1),
(SELECT count(*) FROM platform.domain_cutovers WHERE scope_id=$1),
(SELECT count(*) FROM platform.activation_authority_calls WHERE tenant_id=$1)`,
		fixture.tenantID).Scan(&tenants, &domains, &authorities)
	if err != nil || tenants != 0 || domains != 0 || authorities != 0 {
		t.Fatalf("wrong profile leaked tenant=%d domains=%d authorities=%d err=%v",
			tenants, domains, authorities, err)
	}
}

func assertActivationAndReplay(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture activationScopeFixture, application string,
) {
	t.Helper()
	replayed, err := callActivationV4(t, ctx, database, fixture, application)
	if err != nil || replayed {
		t.Fatalf("%s initial activation replay=%v err=%v", fixture.kind, replayed, err)
	}
	replayed, err = callActivationV4(t, ctx, database, fixture, application)
	if err != nil || !replayed {
		t.Fatalf("%s replay replay=%v err=%v", fixture.kind, replayed, err)
	}
	var tenants, domains, groups, members, authorities, facts int
	err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM platform.tenants WHERE id=$1),
(SELECT count(*) FROM platform.domain_cutovers WHERE scope_id=$1),
(SELECT count(*) FROM platform.cutover_groups WHERE scope_id=$1),
(SELECT count(*) FROM platform.cutover_group_members WHERE scope_id=$1),
(SELECT count(*) FROM platform.activation_authority_calls WHERE tenant_id=$1),
(SELECT count(*) FROM platform.tenancy_activation_facts WHERE tenant_id=$1)`,
		fixture.tenantID).Scan(
		&tenants, &domains, &groups, &members, &authorities, &facts,
	)
	if err != nil || tenants != 1 || domains != 10 || groups != 8 ||
		members != 10 || authorities != 2 || facts != 1 {
		t.Fatalf(
			"%s activation counts tenant=%d domain=%d group=%d member=%d authority=%d fact=%d err=%v",
			fixture.kind, tenants, domains, groups, members, authorities, facts, err,
		)
	}
}

func activationErrorDenied(err error) bool {
	return err != nil && strings.Contains(err.Error(), "initialization v4 denied")
}

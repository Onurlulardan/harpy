//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func assertIAMRefreshLocatorGuards(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) []byte {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	issued, projected := now.Add(-time.Minute), now
	familyExpiry, retain := now.Add(48*time.Hour), now.Add(73*time.Hour)
	selector := iamBytes(17, 16)
	if err := createIAMGuardLocator(ctx, database.runtime, selector,
		iamGuardTenantA, 1, "active", issued, projected, familyExpiry, retain); err != nil {
		t.Fatal(err)
	}
	revision, err := advanceIAMGuardLocator(ctx, database.runtime, selector,
		iamGuardTenantA, 1, 1, 2, "consumed", projected.Add(time.Minute),
		retain.Add(time.Hour))
	if err != nil || revision != 2 {
		t.Fatalf("advance locator revision=%d err=%v", revision, err)
	}
	_, err = advanceIAMGuardLocator(ctx, database.runtime, selector,
		iamGuardTenantA, 1, 1, 3, "consumed", projected.Add(2*time.Minute),
		retain.Add(2*time.Hour))
	assertIAMSQLState(t, err, "P5711")
	_, err = advanceIAMGuardLocator(ctx, database.runtime, selector,
		iamGuardTenantA, 1, 2, 3, "active", projected.Add(2*time.Minute),
		retain.Add(2*time.Hour))
	assertIAMSQLState(t, err, "55000")
	_, err = advanceIAMGuardLocator(ctx, database.runtime, selector,
		iamGuardTenantA, 1, 2, 3, "consumed", projected.Add(2*time.Minute),
		retain.Add(-time.Hour))
	assertIAMSQLState(t, err, "55000")
	assertIAMLocatorAuthorityFailures(t, ctx, database, now)
	assertIAMLocatorWrongSession(t, ctx, database, now)
	return selector
}

func assertIAMLocatorAuthorityFailures(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, now time.Time,
) {
	t.Helper()
	create := func(selector byte, tenant string, generation int64) error {
		return createIAMGuardLocator(ctx, database.runtime, iamBytes(selector, 16),
			tenant, generation, "active", now.Add(-time.Minute), now,
			now.Add(48*time.Hour), now.Add(73*time.Hour))
	}
	assertIAMSQLState(t, create(18, iamGuardTenantA, 2), "P5701")
	setIAMGuardAuthorityState(t, ctx, database, "cutover_groups", "", "v1")
	assertIAMSQLState(t, create(19, iamGuardTenantB, 1), "P5701")
	setIAMGuardAuthorityState(t, ctx, database, "cutover_groups", "", "v2")
	setIAMGuardAuthorityState(t, ctx, database, "domain_cutovers", "iam", "v1")
	assertIAMSQLState(t, create(20, iamGuardTenantB, 1), "P5702")
	setIAMGuardAuthorityState(t, ctx, database, "domain_cutovers", "iam", "v2")
	setIAMGuardAuthorityState(t, ctx, database,
		"domain_cutovers", "system_resource_layout", "v1")
	assertIAMSQLState(t, create(21, iamGuardTenantB, 1), "P5703")
	setIAMGuardAuthorityState(t, ctx, database,
		"domain_cutovers", "system_resource_layout", "v2")
}

func setIAMGuardAuthorityState(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	table, domain, state string,
) {
	t.Helper()
	var statement string
	var arguments []any
	if table == "cutover_groups" {
		statement = `UPDATE platform.cutover_groups SET state=$2,writer=$2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='iam_system_resources'`
		arguments = []any{iamGuardTenantB, state}
	} else {
		statement = `UPDATE platform.domain_cutovers SET state=$3,writer=$3
WHERE scope_kind='tenant' AND scope_id=$1 AND domain=$2`
		arguments = []any{iamGuardTenantB, domain, state}
	}
	if _, err := database.migrator.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatal(err)
	}
}

func assertIAMLocatorWrongSession(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase, now time.Time,
) {
	t.Helper()
	err := createIAMGuardLocator(ctx, database.migrator, iamBytes(22, 16),
		iamGuardTenantA, 1, "active", now.Add(-time.Minute), now,
		now.Add(48*time.Hour), now.Add(73*time.Hour))
	assertIAMSQLState(t, err, "42501")
}

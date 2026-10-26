//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

func assertMaintenancePG18ProfileCapabilityACL(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const routines = `
SELECT pg_catalog.count(*),pg_catalog.bool_and(routine.prosecdef),
 pg_catalog.bool_and(owner.rolname='harpy_owner'),
 pg_catalog.bool_and(routine.proconfig=ARRAY[
  'search_path='||$1||', pg_catalog']::text[]),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'harpy_runtime',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'harpy_adopter',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'public',routine.oid,'EXECUTE'))
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
 'iam_profile_worker_capability_mint_v1',
 'iam_profile_worker_capability_consume_v1',
 'iam_profile_worker_capability_commit_v1',
	'iam_guard_profile_backfill_source_write_v3',
	'iam_guard_profile_backfill_source_write_v4'])`
	var count int
	var definer, owner, captured, runtime, adopter, public bool
	err := database.QueryRowContext(ctx, routines, schema).Scan(
		&count, &definer, &owner, &captured, &runtime, &adopter, &public,
	)
	if err != nil || count != 5 || !definer || !owner || !captured ||
		!runtime || !adopter || !public {
		t.Fatalf("%s capability ACL=%d/%t/%t/%t/%t/%t/%t: %v",
			schema, count, definer, owner, captured, runtime, adopter, public, err)
	}
	const triggers = `
SELECT pg_catalog.count(*) FILTER (WHERE
  (relation.relname=ANY(ARRAY['iam_users','iam_organizations','iam_roles'])
   AND routine.proname='iam_guard_profile_backfill_source_write_v4')
  OR (relation.relname='iam_profile_backfill_worker_capabilities'
   AND routine.proname='iam_profile_worker_capability_commit_v1')),
 pg_catalog.count(*) FILTER (WHERE
  routine.proname='iam_guard_profile_backfill_source_write_v3')
FROM pg_catalog.pg_trigger AS trigger
JOIN pg_catalog.pg_class AS relation ON relation.oid=trigger.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
JOIN pg_catalog.pg_proc AS routine ON routine.oid=trigger.tgfoid
WHERE namespace.nspname=$1 AND NOT trigger.tgisinternal
 AND trigger.tgenabled='A'`
	var triggerCount, legacyTriggerCount int
	err = database.QueryRowContext(ctx, triggers, schema).Scan(
		&triggerCount, &legacyTriggerCount,
	)
	if err != nil || triggerCount != 7 || legacyTriggerCount != 0 {
		t.Fatalf("%s capability triggers=%d legacy=%d: %v",
			schema, triggerCount, legacyTriggerCount, err)
	}
}

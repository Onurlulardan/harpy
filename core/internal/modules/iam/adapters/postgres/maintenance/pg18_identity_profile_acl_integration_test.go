//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

func assertMaintenancePG18IdentityProfileACL(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const statement = `
SELECT pg_catalog.count(*),pg_catalog.bool_and(routine.prosecdef),
 pg_catalog.bool_and(owner.rolname='harpy_owner'),
 pg_catalog.bool_and(routine.proconfig=ARRAY[
  'search_path='||$1||', pg_catalog']::text[]),
 pg_catalog.bool_and(pg_catalog.has_function_privilege(
   'harpy_runtime',routine.oid,'EXECUTE') = (
    routine.proname IN ('iam_identity_organization_create_v3',
      'iam_identity_organization_update_v3') OR
    (pg_catalog.right(routine.proname,3)='_v2' AND routine.proname NOT IN (
      'iam_identity_organization_create_v2','iam_identity_organization_update_v2'))
  )),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'harpy_adopter',routine.oid,'EXECUTE')),
 pg_catalog.bool_and(NOT pg_catalog.has_function_privilege(
  'public',routine.oid,'EXECUTE'))
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE namespace.nspname=$1 AND routine.proname=ANY(ARRAY[
 'iam_identity_user_create_v1','iam_identity_user_create_v2',
 'iam_identity_user_update_v1','iam_identity_user_update_v2',
 'iam_identity_organization_create_v1','iam_identity_organization_create_v2',
 'iam_identity_organization_create_v3','iam_identity_organization_update_v1',
 'iam_identity_organization_update_v2','iam_identity_organization_update_v3',
 'iam_identity_role_create_v1','iam_identity_role_create_v2',
 'iam_identity_role_update_v1','iam_identity_role_update_v2',
 'iam_invitation_credential_lookup_v1','iam_invitation_credential_lookup_v2',
 'iam_invitation_register_accept_v1','iam_invitation_register_accept_v2',
 'iam_self_profile_update_v1','iam_self_profile_update_v2'])`
	var count int
	var definer, owner, captured, runtime, adopter, public bool
	err := database.QueryRowContext(ctx, statement, schema).Scan(
		&count, &definer, &owner, &captured, &runtime, &adopter, &public,
	)
	if err != nil || count != 20 || !definer || !owner || !captured ||
		!runtime || !adopter || !public {
		t.Fatalf("%s identity profile ACL=%d/%t/%t/%t/%t/%t/%t: %v",
			schema, count, definer, owner, captured, runtime, adopter, public, err)
	}
}

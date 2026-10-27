//go:build integration

package postgres

import (
	"context"
	"testing"
)

func assertInstallationOwnerVerifierACLPG18(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
) {
	t.Helper()
	var owner, definer, path, runtime, login, adopter, restore bool
	err := fixture.migrator.QueryRowContext(ctx, `
SELECT routine.proowner=to_regrole('harpy_owner'),routine.prosecdef,
       routine.proconfig @> ARRAY['search_path=pg_catalog']::text[],
       has_function_privilege('harpy_runtime',routine.oid,'EXECUTE'),
       has_function_privilege('harpy_v2_runtime',routine.oid,'EXECUTE'),
       has_function_privilege('harpy_adopter',routine.oid,'EXECUTE'),
       has_function_privilege('harpy_restore',routine.oid,'EXECUTE')
FROM pg_catalog.pg_proc AS routine
WHERE routine.oid=to_regprocedure(
 'platform.iam_installation_instance_owner_verify_v1(uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text)'
)`).Scan(&owner, &definer, &path, &runtime, &login, &adopter, &restore)
	if err != nil {
		t.Fatal(err)
	}
	if !owner || !definer || !path || !runtime || login || adopter || restore {
		t.Fatalf("installation owner verifier ACL = %t/%t/%t/%t/%t/%t/%t",
			owner, definer, path, runtime, login, adopter, restore)
	}
	_, err = callInstallationOwnerVerifierPG18(ctx, fixture.runtime, ownerVerifierRequest)
	if err != nil {
		t.Fatalf("runtime ACL rejected verifier: %v", err)
	}
}

func installationOwnerVerifierSnapshotPG18(
	t *testing.T, ctx context.Context, fixture ownerTransferFixture,
) string {
	t.Helper()
	var snapshot string
	err := fixture.migrator.QueryRowContext(ctx, `
SELECT jsonb_build_object(
 'subjects',(SELECT jsonb_agg(to_jsonb(row) ORDER BY id) FROM platform.iam_platform_subjects AS row WHERE environment_id=$1),
 'roles',(SELECT jsonb_agg(to_jsonb(row) ORDER BY id) FROM platform.iam_platform_roles AS row WHERE environment_id=$1),
 'grants',(SELECT jsonb_agg(to_jsonb(row) ORDER BY id) FROM platform.iam_platform_role_permissions AS row WHERE environment_id=$1),
 'bindings',(SELECT jsonb_agg(to_jsonb(row) ORDER BY id) FROM platform.iam_platform_subject_roles AS row WHERE environment_id=$1),
 'policy',(SELECT to_jsonb(row) FROM platform.iam_platform_owner_policies AS row WHERE environment_id=$1),
 'clock',(SELECT to_jsonb(row) FROM platform.iam_platform_authorization_clocks AS row WHERE environment_id=$1),
 'receipts',(SELECT jsonb_agg(to_jsonb(row) ORDER BY operation_id) FROM platform.iam_platform_mutation_receipts AS row WHERE environment_id=$1)
)::text`, ownerEnvID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

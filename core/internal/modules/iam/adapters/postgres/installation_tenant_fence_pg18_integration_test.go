//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func applyIAMTenantCapabilityFencePG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string,
) {
	t.Helper()
	connection, err := database.migrator.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	_, err = connection.ExecContext(ctx, `
REVOKE ALL ON FUNCTION
    platform.tenancy_activation_step_context_assert_v1(
        uuid,text,uuid,uuid,name,text
    ),
    platform.tenancy_activation_step_consume_v1(uuid,text)
FROM PUBLIC, harpy_runtime, harpy_v2_runtime;
CREATE SCHEMA iam_capability AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA iam_capability TO harpy_runtime;
SELECT pg_catalog.set_config(
    'search_path', 'iam_capability,pg_catalog', false
);
CREATE TABLE test_identity_writes (
    operation_id uuid PRIMARY KEY
);
CREATE FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER
SET search_path FROM CURRENT AS $body$
DECLARE v_inserted integer;
BEGIN
    INSERT INTO test_identity_writes VALUES ($1) ON CONFLICT DO NOTHING;
    GET DIAGNOSTICS v_inserted = ROW_COUNT;
    RETURN v_inserted = 0;
END
$body$;
REVOKE ALL ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC, harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) TO harpy_runtime;
DO $fixture$
DECLARE
    v_signature text;
BEGIN
    FOREACH v_signature IN ARRAY ARRAY[
        'iam_installation_credential_initialize_v1(name,uuid,uuid,uuid,uuid,bigint,bytea,bytea)',
        'iam_initial_admin_setup_finalize_v1(name,uuid,uuid,text)'
    ] LOOP
        EXECUTE pg_catalog.format(
            'CREATE FUNCTION %s RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS %L',
            v_signature, 'BEGIN RETURN; END'
        );
        EXECUTE pg_catalog.format(
            'REVOKE ALL ON FUNCTION %s FROM PUBLIC, harpy_v2_runtime',
            v_signature
        );
        EXECUTE pg_catalog.format(
            'GRANT EXECUTE ON FUNCTION %s TO harpy_runtime', v_signature
        );
    END LOOP;
END
$fixture$;`)
	if err != nil {
		t.Fatalf("prepare IAM tenant fence fixture: %v", err)
	}
	for _, name := range []string{
		"20260729100025_create_iam_installation_tenant_capability_entrypoints_v1",
		"20260729100035_assert_iam_installation_capabilities_v1",
	} {
		document, readErr := os.ReadFile(filepath.Join(
			"..", "..", "migrations", "tenant", origin,
			name+"."+origin+".sql",
		))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, applyErr := connection.ExecContext(
			ctx, string(document),
		); applyErr != nil {
			t.Fatalf("apply IAM tenant capability migration %s: %v",
				name, applyErr)
		}
	}
	if _, err := database.runtime.ExecContext(ctx,
		"SET search_path TO pg_catalog, iam_capability"); err != nil {
		t.Fatalf("bind IAM tenant runtime search path: %v", err)
	}
}

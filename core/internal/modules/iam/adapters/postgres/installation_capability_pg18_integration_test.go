//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallationCapabilitySuccessorsOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "install_cap_"+origin)
			prepareInstallationCapabilityPG18(t, ctx, database, origin)
			assertInstallationNoCapabilityPG18(t, ctx, database)
			assertInstallationLegacyAuthorityPG18(t, ctx, database)
			applyInstallationAuthorityContractsPG18(t, ctx, database, origin)
			assertInstallationContractedAuthorityPG18(t, ctx, database)
			assertInstallationOwnerCapabilityPG18(t, ctx, database)
			assertInstallationAsyncEvidencePG18(t, ctx, database)
		})
	}
}

func prepareInstallationCapabilityPG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	origin string,
) {
	t.Helper()
	for _, fixtureName := range []string{
		"installation_capability_pg18.sqlfixture",
		"installation_authority_contract_pg18.sqlfixture",
	} {
		fixture, err := os.ReadFile(filepath.Join("testdata", fixtureName))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := database.migrator.ExecContext(ctx, string(fixture)); err != nil {
			t.Fatalf("prepare %s: %v", fixtureName, err)
		}
	}
	for _, path := range []string{
		filepath.Join("..", "..", "migrations", "platform", origin,
			"20260729100024_create_iam_installation_capability_entrypoints_v1."+
				origin+".sql"),
		filepath.Join("..", "..", "..", "..", "platform", "async", "migrations",
			origin,
			"20260729100029_create_installation_activation_capability_entrypoints_v1."+
				origin+".sql"),
	} {
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, applyErr := database.migrator.ExecContext(ctx, string(body)); applyErr != nil {
			t.Fatalf("apply %s: %v", path, applyErr)
		}
	}
	applyIAMTenantCapabilityFencePG18(t, ctx, database, origin)
	_, err := database.migrator.ExecContext(ctx, `
CREATE FUNCTION platform.test_activation_evidence_proxy(
    p_operation_id uuid,p_expected_artifact_id uuid
) RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $$
BEGIN
    RETURN platform.platform_async_installation_activation_evidence_assert_v1(
        p_operation_id,p_expected_artifact_id
    );
END
$$;
REVOKE ALL ON FUNCTION platform.test_activation_evidence_proxy(uuid,uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.test_activation_evidence_proxy(uuid,uuid)
TO harpy_runtime;`)
	if err != nil {
		t.Fatalf("create activation evidence proxy: %v", err)
	}
	if _, err := database.runtime.ExecContext(ctx,
		"SET application_name TO 'harpy-v2:core-worker'"); err != nil {
		t.Fatal(err)
	}
}

func assertInstallationApplicationNamePG18(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	var applicationName string
	if err := database.runtime.QueryRowContext(ctx,
		"SELECT current_setting('application_name')").Scan(&applicationName); err != nil {
		t.Fatal(err)
	}
	if applicationName != "harpy-v2:core-worker" {
		t.Fatalf("application_name leaked as %q", applicationName)
	}
}

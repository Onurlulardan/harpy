//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGen70ActivationFenceAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), 3*time.Minute,
			)
			defer cancel()
			database := startCutoverIntegrationDatabaseForOriginSkipping(
				t, ctx, origin, map[string]bool{
					"20260728110020_assert_runtime_snapshot_authorities_v1." +
						origin + ".sql": true,
					"20260729100045_assert_gen71_activation_contract_v1." +
						origin + ".sql": true,
				},
			)
			document, err := os.ReadFile(filepath.Join(
				"..", "..", "migrations", origin,
				"20260729100034_assert_gen70_activation_capabilities_v1."+
					origin+".sql",
			))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = database.admin.Exec(ctx, string(document)); err != nil {
				t.Fatalf("reapply healthy Gen70 activation fence: %v", err)
			}
			if _, err = database.admin.Exec(ctx, `
REVOKE EXECUTE ON FUNCTION
platform.billing_installation_assign_v3(
    uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,
    bigint,text,text,text
) FROM harpy_runtime`); err != nil {
				t.Fatal(err)
			}
			if _, err = database.admin.Exec(ctx, string(document)); err == nil {
				t.Fatal("Gen70 activation fence accepted a closed public successor")
			}
		})
	}
}

func TestGen71ActivationContractAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(), 3*time.Minute,
			)
			defer cancel()
			database := startCutoverIntegrationDatabaseForOriginSkipping(
				t, ctx, origin, map[string]bool{
					"20260728110020_assert_runtime_snapshot_authorities_v1." +
						origin + ".sql": true,
					"20260729100045_assert_gen71_activation_contract_v1." +
						origin + ".sql": true,
				},
			)
			applyGen71OwnerContracts(t, ctx, database.admin, origin)
			document, err := os.ReadFile(filepath.Join(
				"..", "..", "migrations", origin,
				"20260729100045_assert_gen71_activation_contract_v1."+
					origin+".sql",
			))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = database.admin.Exec(ctx, string(document)); err != nil {
				t.Fatalf("reapply healthy Gen71 activation contract: %v", err)
			}
			if _, err = database.admin.Exec(ctx, `
GRANT EXECUTE ON FUNCTION
platform.tenancy_onboarding_activate_v2(uuid,text,bigint)
TO harpy_runtime`); err != nil {
				t.Fatal(err)
			}
			if _, err = database.admin.Exec(ctx, string(document)); err == nil {
				t.Fatal("Gen71 activation contract accepted retired runtime ABI")
			}
		})
	}
}

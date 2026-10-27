//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMProfileKeyRepairUpgradeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(
				t, ctx, "profile_key_repair_"+origin,
			)
			fixtures := prepareIAMProfileKeyRepairPG18(
				t, ctx, database, origin,
			)
			assertIAMProfileRepairRejectsCorruptionPG18(
				t, ctx, database, origin, fixtures,
			)
			applyIAMProfileRepairTargetsPG18(
				t, ctx, database.migrator, origin,
				fixtures.legacy.schema, fixtures.rolling.schema,
			)
			assertIAMProfileRepairRowsPG18(
				t, ctx, database.migrator, fixtures.legacy, 2,
			)
			applyIAMIdentityParityInstallTargetsPG18(
				t, ctx, database.migrator, origin,
				fixtures.legacy.schema, fixtures.rolling.schema,
			)
			assertIAMInstallationUserProfileV2PG18(
				t, ctx, database.migrator, fixtures.legacy, 3,
			)
			replayed, err := callIAMInstallationPG18(
				ctx, database.runtime, fixtures.legacy, "v4",
			)
			if err != nil || !replayed {
				t.Fatalf("v4 installation replay = %t: %v", replayed, err)
			}
			assertIAMInstallationUserProfileV2PG18(
				t, ctx, database.migrator, fixtures.legacy, 3,
			)
			replayed, err = callIAMInstallationPG18(
				ctx, database.runtime, fixtures.rolling, "v4",
			)
			if err != nil || replayed {
				t.Fatalf("rolling v4 installation = %t: %v", replayed, err)
			}
			assertIAMInstallationUserProfileV2PG18(
				t, ctx, database.migrator, fixtures.rolling, 2,
			)
			replayed, err = callIAMInstallationPG18(
				ctx, database.runtime, fixtures.rolling, "v4",
			)
			if err != nil || !replayed {
				t.Fatalf("rolling v4 replay = %t: %v", replayed, err)
			}
			assertIAMInstallationUserProfileV2PG18(
				t, ctx, database.migrator, fixtures.rolling, 2,
			)
		})
	}
}

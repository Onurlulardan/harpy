//go:build integration

package maintenance

import (
	"context"
	"testing"
	"time"
)

func TestMaintenanceRuntimeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startMaintenancePG18(t, ctx, origin)
			prepareMaintenancePG18Schemas(t, ctx, database, origin)
			applyMaintenancePG18Platform(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "platform",
					maintenanceFoundationVersion("platform"), 3))
			applyMaintenancePG18Tenant(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "tenant",
					maintenanceFoundationVersion("tenant"), 23))
			seedMaintenancePG18Foundation(t, ctx, database)
			applyMaintenancePG18Tenant(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "tenant",
					maintenanceGuardVersion, 7))
			applyMaintenancePG18Tenant(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "tenant",
					maintenancePrerequisiteVersion, 38))
			seedMaintenancePG18Authority(t, ctx, database)
			seedMaintenancePG18ProfileSources(t, ctx, database)
			seedMaintenancePG18OptionalProfileVersion(t, ctx, database)
			prepareMaintenancePG18Runtime(t, ctx, database)
			applyMaintenancePG18Platform(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "platform",
					maintenancePlatformVersion, 4))
			applyMaintenancePG18Tenant(t, ctx, database,
				loadMaintenancePG18Migrations(t, origin, "tenant",
					maintenanceTenantVersion, 121))
			assertMaintenancePG18Catalog(t, ctx, database)
			assertMaintenancePG18ACL(t, ctx, database)
			assertMaintenancePG18DirectRuntime(t, ctx, database)
			assertMaintenancePG18DirectoryExecution(t, ctx, database)
			assertMaintenancePG18ProfileExecution(t, ctx, database)
		})
	}
}

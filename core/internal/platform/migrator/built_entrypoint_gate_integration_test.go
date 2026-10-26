//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestBuiltEntrypointsAgainstCanonicalPostgreSQL184Contract(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	stage := prepareBuiltEntrypointStage(t, ctx)
	binaries := buildEntrypointBinaries(t, ctx, stage)
	database := startBuiltEntrypointDatabase(t, ctx)
	evidence := migrateBuiltEntrypointDatabase(
		t, ctx, database, stage, binaries.migrator,
	)
	profiles := []builtEntrypointProfile{
		{name: "api", binary: binaries.api, dsn: database.runtimeDSN,
			service: runtimev1.ServiceCoreAPI, sessionUser: "harpy_v2_runtime",
			application: "harpy-v2:core-api", api: true,
			catalogPackageKeyring: evidence.catalog.catalogPackageKeyring},
		{name: "worker", binary: binaries.worker, dsn: database.runtimeDSN,
			service: runtimev1.ServiceCoreWorker, sessionUser: "harpy_v2_runtime",
			application:           "harpy-v2:core-worker",
			catalogPackageKeyring: evidence.catalog.catalogPackageKeyring},
		{name: "migrator", binary: binaries.migrator,
			arguments: []string{"schema-executor"}, dsn: database.migratorDSN,
			service: runtimev1.ServiceCoreMigrator, sessionUser: "harpy_migrator",
			application:           "harpy-v2:core-schema-executor",
			catalogPackageKeyring: evidence.catalog.catalogPackageKeyring},
	}
	for _, profile := range profiles {
		profile := profile
		t.Run(profile.name, func(t *testing.T) {
			runBuiltEntrypoint(t, ctx, database, profile)
		})
	}
}

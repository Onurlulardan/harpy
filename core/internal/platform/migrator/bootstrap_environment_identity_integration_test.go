//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"path"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const globalBillingAuthorityVersion = int64(20260716020103)

func TestRunnerPersistsIdentityBeforeGlobalBillingAuthorityOnPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []postgresv1.DatabaseOrigin{postgresv1.DatabaseOriginFresh} {
		t.Run(string(origin), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database, config := startRunnerIntegrationDatabase(t, ctx)
			if origin == postgresv1.DatabaseOriginV1Adopted {
				prepareRunnerV1TenantRegistry(t, ctx, database)
			}
			runner := newRunnerIntegrationFixture(t, ctx, database, config)
			files, selected := runnerIntegrationCatalogForOrigin(t, origin)
			runner.platformFiles, runner.selected = platformCatalogThrough(
				t, files, selected, globalBillingAuthorityVersion,
			)
			runner.identity.databaseOrigin = origin
			last := runner.selected[len(runner.selected)-1]
			runner.identity.databaseGeneration = last.TargetDatabaseGeneration
			runner.identity.minimumRuntimeGeneration = last.NextMinimumRuntimeGeneration
			runID := runnerIntegrationRunID
			if origin == postgresv1.DatabaseOriginV1Adopted {
				runID = "018f0000-0000-7000-8000-000000000113"
			}
			provider, err := runner.buildGooseProvider(runID)
			if err != nil {
				t.Fatal(err)
			}
			connection, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close()
			if err := runner.bootstrapRunLifecycle(ctx, connection, provider); err != nil {
				t.Fatalf("bootstrap migration lifecycle: %v", err)
			}
			if _, err := runner.ensureRunRunning(ctx, connection, runID); err != nil {
				t.Fatalf("start migration run: %v", err)
			}
			if _, err := provider.UpTo(ctx, globalBillingAuthorityVersion); err != nil {
				t.Fatalf("apply through global billing authority: %v", err)
			}
			assertGlobalBillingBootstrapIdentity(t, ctx, database, origin)
		})
	}
}

func platformCatalogThrough(
	t *testing.T,
	files fstest.MapFS,
	selected []postgresv1.SelectedMigration,
	maxVersion int64,
) (fstest.MapFS, []postgresv1.SelectedMigration) {
	t.Helper()
	filteredFiles := fstest.MapFS{}
	filtered := make([]postgresv1.SelectedMigration, 0, len(selected))
	for _, migration := range selected {
		if migration.Version > maxVersion {
			continue
		}
		name := path.Base(migration.Path)
		version, err := strconv.ParseInt(name[:14], 10, 64)
		if err != nil || version != migration.Version {
			t.Fatalf("parse platform migration %s: %v", name, err)
		}
		filteredFiles[name] = files[name]
		filtered = append(filtered, migration)
	}
	return filteredFiles, filtered
}

func assertGlobalBillingBootstrapIdentity(
	t *testing.T,
	ctx context.Context,
	database interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	},
	origin postgresv1.DatabaseOrigin,
) {
	t.Helper()
	state := "v2"
	if origin == postgresv1.DatabaseOriginV1Adopted {
		state = "v1"
	}
	var identityCount, domainCount, groupCount, memberCount int
	err := database.QueryRowContext(ctx, `SELECT
    (SELECT count(*) FROM platform.environment_identity
      WHERE singleton AND database_origin=$1),
    (SELECT count(*) FROM platform.domain_cutovers
      WHERE scope_kind='global' AND domain='billing' AND state=$2 AND writer=$2),
    (SELECT count(*) FROM platform.cutover_groups
      WHERE scope_kind='global' AND group_id='billing' AND state=$2 AND writer=$2),
    (SELECT count(*) FROM platform.cutover_group_members
      WHERE scope_kind='global' AND group_id='billing' AND ordinal=1 AND domain='billing')`,
		string(origin), state,
	).Scan(&identityCount, &domainCount, &groupCount, &memberCount)
	if err != nil || identityCount != 1 || domainCount != 1 ||
		groupCount != 1 || memberCount != 1 {
		t.Fatalf("bootstrap identity/billing counts = %d/%d/%d/%d err=%v",
			identityCount, domainCount, groupCount, memberCount, err)
	}
}

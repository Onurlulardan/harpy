//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const publicationCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

type publicationRuntime struct {
	pool        *platformpostgres.Pool
	coordinator *platformtx.Coordinator
	application iammodule.ProfileHTTPApplication
	recoveries  *iammodule.ProfileBackfillRecoveryRuntime
}

type publicationReferences struct{}

func (publicationReferences) AssertRetirable(
	context.Context, platformtx.TenantScope, profileapp.RetirementTarget,
) (profileapp.ReferenceEvidence, error) {
	return profileapp.ReferenceEvidence{}, profileapp.ErrEvidenceMissing
}

type publicationChanges struct{}

func (publicationChanges) WriteChange(
	context.Context, platformtx.TenantScope, changeapp.Change,
) error {
	return nil
}

func openPublicationRuntime(
	t *testing.T, ctx context.Context, db publicationDatabase,
) publicationRuntime {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", db.runtimeDSN)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", publicationCursorKeyring)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "4")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := platformpostgres.Open(
		ctx, loaded.LongRunning().Runtime().Postgres(), platformpostgres.ProfileAPI,
	)
	if err != nil {
		t.Fatalf("open publication runtime pool: %v", err)
	}
	t.Cleanup(pool.Close)
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		t.Fatal(err)
	}
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	backfills, err := iammodule.NewProfileBackfillPublicationRuntime(
		catalogs, func() time.Time {
			return time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := iammodule.NewProfileCommands(publicationReferences{})
	if err != nil {
		t.Fatal(err)
	}
	recoveries, err := iammodule.NewProfileBackfillRecoveryRuntime(
		catalogs, func() time.Time {
			return time.Date(2026, 7, 16, 17, 0, 0, 0, time.UTC)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	application, err := iammodule.NewProfileHTTPApplication(
		pool, coordinator, commands, backfills, recoveries, publicationChanges{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return publicationRuntime{
		pool: pool, coordinator: coordinator, application: application,
		recoveries: recoveries,
	}
}

func grantPublicationRuntimeAccess(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const sql = `SET ROLE harpy_owner;
GRANT USAGE ON SCHEMA platform,t_profile_publication TO harpy_runtime;
GRANT SELECT ON platform.tenants,platform.domain_cutovers,
 platform.cutover_groups,platform.cutover_group_members,
 platform.environment_identity,platform.schema_contracts,
 platform.database_generation_manifests,platform.release_generations,
 platform.migration_runs,platform.bootstrap_receipts,
 platform.extension_registry TO harpy_runtime;
GRANT UPDATE ON platform.tenants,platform.cutover_groups,
 platform.domain_cutovers TO harpy_runtime;
GRANT SELECT ON t_profile_publication.tenancy_schema_identity TO harpy_runtime;
RESET ROLE;`
	if _, err := db.admin.Exec(ctx, sql); err != nil {
		t.Fatalf("grant publication runtime access: %v", err)
	}
}

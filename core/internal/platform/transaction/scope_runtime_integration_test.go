//go:build integration

package transaction

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func openScopePool(t *testing.T, ctx context.Context, dsn string) *postgres.Pool {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", dsn)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", scopeTestCursorKeyring)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "2")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Open(
		ctx, loaded.LongRunning().Runtime().Postgres(), postgres.ProfileAPI,
		scopeSchemaContract(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func scopeSchemaContract(t *testing.T) schemacontract.EmbeddedContract {
	t.Helper()
	migrationDigest := scopeMigrationDigest(t)
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion, Component: "core",
		RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: migrationDigest,
		ExtensionManifestSHA256: scopeExtensionDigest,
		PostgresImageDigest:     runtimev1.PostgresImageDigest,
	}
	document, err := attestation.MarshalCanonical(contract)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := binarycontract.Verify(document)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := schemacontract.AuthorizeIntegrationArtifact(registry, artifact)
	if err != nil {
		t.Fatal(err)
	}
	return authorized
}

func scopeMigrationDigest(t *testing.T) string {
	t.Helper()
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	digests := registry.Digests()
	if len(digests) == 0 {
		t.Fatal("migration release registry is empty")
	}
	return digests[0]
}

const scopeExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const scopeTestCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func scopeTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 5*time.Minute)
}

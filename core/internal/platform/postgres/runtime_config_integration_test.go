//go:build integration

package postgres

import (
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func runtimePoolConfig(t *testing.T, dsn string) config.Postgres {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", dsn)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", runtimeTestCursorKeyring)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "2")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatalf("load runtime pool config: %v", err)
	}
	return loaded.LongRunning().Runtime().Postgres()
}

func runtimeSchemaContract(t *testing.T) schemacontract.EmbeddedContract {
	t.Helper()
	return runtimeSchemaContractFor(t, 1, 1, runtimeMigrationDigest(t))
}

func runtimeSchemaContractFor(
	t *testing.T,
	runtimeGeneration int64,
	minimumDatabaseGeneration int64,
	migrationDigest string,
) schemacontract.EmbeddedContract {
	t.Helper()
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion, Component: "core",
		RuntimeGeneration:         runtimeGeneration,
		MinimumDatabaseGeneration: minimumDatabaseGeneration,
		MigrationManifestSHA256:   migrationDigest,
		ExtensionManifestSHA256:   runtimeExtensionDigest,
		PostgresImageDigest:       runtimev1.PostgresImageDigest,
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

func runtimeMigrationDigest(t *testing.T) string {
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

func runtimeMigrationDigests(t *testing.T) []string {
	t.Helper()
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	return registry.Digests()
}

const runtimeExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const runtimeTestCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

//go:build integration

package executor

import (
	"context"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

const pg18CursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func pg18OpenRuntime(
	t *testing.T, ctx context.Context, dsn string,
) (*postgres.Pool, cursor.Codec) {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", dsn)
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", pg18CursorKeyring)
	t.Setenv("HARPY_POSTGRES_MAX_CONNS", "2")
	t.Setenv("HARPY_POSTGRES_MIN_CONNS", "0")
	t.Setenv("HARPY_POSTGRES_CLEANUP_TIMEOUT", "5s")
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Open(
		ctx, loaded.LongRunning().Runtime().Postgres(),
		postgres.ProfileAPI, pg18SchemaContract(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	keyring, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.NewCodec(keyring)
	if err != nil {
		t.Fatal(err)
	}
	return pool, codec
}

func pg18SchemaContract(t *testing.T) schemacontract.EmbeddedContract {
	t.Helper()
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1,
		MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256:   pg18MigrationDigest(t),
		ExtensionManifestSHA256:   pg18ExtensionDigest,
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

func pg18MigrationDigest(t *testing.T) string {
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

func pg18ScopeRequest(
	t *testing.T, tenant string, owners ...string,
) transaction.ScopeRequest {
	t.Helper()
	actor, err := transaction.NewActorIssuer().Issue(
		uuid.MustParse(pg18Actor), uuid.MustParse(tenant),
	)
	if err != nil {
		t.Fatal(err)
	}
	generations := make(map[string]int64, len(owners))
	for _, owner := range owners {
		generations[owner] = 1
	}
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotOperationID, generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	return transaction.ScopeRequest{
		Actor: actor, Operation: operation, Mode: transaction.SnapshotQuery,
	}
}

//go:build integration

package migrator

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"log/slog"
	"testing"
	"time"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/jackc/pgx/v5"
)

const (
	runnerIntegrationReleaseID     = "018f0000-0000-7000-8000-000000000101"
	runnerIntegrationEnvironmentID = "018f0000-0000-7000-8000-000000000102"
	runnerIntegrationRunID         = "018f0000-0000-7000-8000-000000000103"
)

type integrationSchemaAuthority struct {
	contract runtimev1.BinarySchemaContract
}

func (authority integrationSchemaAuthority) Contract() runtimev1.BinarySchemaContract {
	return authority.contract
}

func newRunnerIntegrationFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	config *pgx.ConnConfig,
) *Runner {
	t.Helper()
	files, selected := runnerIntegrationCatalog(t)
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("open runner identity connection: %v", err)
	}
	live, err := postgresstore.ReadLiveDatabaseIdentity(ctx, connection)
	_ = connection.Close()
	if err != nil {
		t.Fatalf("read runner integration identity: %v", err)
	}
	migrationHash := integrationDigest("migration")
	extensionHash := integrationDigest("extension")
	last := selected[len(selected)-1]
	runner := &Runner{
		database: database, platformFiles: files, selected: selected,
		tenantDatabase: func() (*sql.DB, error) {
			return openOneShotTenantDatabase(config.Copy()), nil
		},
		identity: runIdentity{
			releaseID:     runnerIntegrationReleaseID,
			environmentID: runnerIntegrationEnvironmentID,
			databaseName:  live.Name, databaseOrigin: postgresv1.DatabaseOriginFresh,
			databaseOwner: live.Owner, databaseEncoding: live.Encoding,
			databaseLocaleProvider: live.LocaleProvider,
			databaseCollation:      live.Collation, databaseCType: live.CType,
			dataChecksums:  live.Checksums,
			manifestSHA256: migrationHash, extensionSHA256: extensionHash,
			rolesSHA256:      integrationDigest("roles"),
			privilegesSHA256: integrationDigest("privileges"),
			ownershipSHA256:  integrationDigest("ownership"),
			authoritySHA256:  integrationDigest("authority"),
			databaseSystemID: live.SystemID, databaseOID: live.OID,
			databaseGeneration:       last.TargetDatabaseGeneration,
			minimumRuntimeGeneration: last.NextMinimumRuntimeGeneration,
		},
		evidence:       runnerIntegrationEvidence(),
		cleanupTimeout: requiredCleanupTimeout,
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	runner.schemaAuthority = integrationSchemaAuthority{contract: runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: migrationHash,
		ExtensionManifestSHA256: extensionHash,
		PostgresImageDigest:     runtimev1.PostgresImageDigest,
	}}
	runner.tenantFiles, runner.tenantSelected = tenantMigrationSources(
		t, postgresv1.DatabaseOriginFresh,
	)
	return runner
}

func runnerIntegrationEvidence() foundationEvidence {
	issuedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	kinds := []string{"cluster_bootstrap", "extension_admin", "credential_cleanup"}
	receipts := make([]receiptEvidence, 0, len(kinds))
	for index, kind := range kinds {
		document := []byte{'{', '"', byte('a' + index), '"', ':', '1', '}'}
		receipts = append(receipts, receiptEvidence{
			kind:              kind,
			runID:             "018f0000-0000-7000-8000-00000000010" + string(rune('4'+index)),
			operationDigest:   integrationDigest(kind),
			canonicalDocument: document,
			canonicalSHA256:   integrationBytesDigest(document),
			signatureKeyID:    "release_key", issuedAt: issuedAt.Add(time.Duration(index) * time.Second),
		})
	}
	return foundationEvidence{receipts: receipts, extensions: []extensionEvidence{}}
}

func integrationDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func integrationBytesDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

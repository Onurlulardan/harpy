//go:build integration

package migrator

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

const rollingReleaseTwoID = "018f0000-0000-7000-8000-000000000201"
const rollingRunTwoID = "018f0000-0000-7000-8000-000000000202"
const rollingRecoveryRunID = "018f0000-0000-7000-8000-000000000299"

func TestRunnerGenerationTwoExpandAndFinalizedReplayOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database, config := startRunnerIntegrationDatabase(t, ctx)
	baseline := newRunnerIntegrationFixture(t, ctx, database, config)
	baselineFiles, baselineSelected := runnerIntegrationBaselineCatalog(t)
	baseline.platformFiles, baseline.selected = baselineFiles, baselineSelected
	baseline.tenantFiles, baseline.tenantSelected = nil, nil
	baseline.identity.databaseGeneration = 1
	baseline.identity.minimumRuntimeGeneration = 1
	if _, err := baseline.Up(ctx, UpRequest{RunID: runnerIntegrationRunID}); err != nil {
		t.Fatalf("apply generation one foundation: %v", err)
	}
	seedRollingUpgradeTenant(t, ctx, database)

	expand := newRunnerIntegrationFixture(t, ctx, database, config)
	configureRollingRelease(
		expand, rollingReleaseTwoID, integrationDigest("migration-generation-two"),
		2, 1, rollingEvidenceTwo(),
	)
	if _, err := expand.Up(ctx, UpRequest{RunID: rollingRunTwoID}); err != nil {
		t.Fatalf("apply generation two expand: %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	identity, err := postgresstore.GetEnvironmentIdentity(ctx, connection)
	if err != nil || identity.BootstrapReleaseID != runnerIntegrationReleaseID {
		t.Fatalf("bootstrap release changed: %#v err=%v", identity, err)
	}
	contract, err := postgresstore.GetCoreSchemaContract(ctx, connection)
	if err != nil || contract.DatabaseGeneration != 2 || contract.Revision != 2 ||
		contract.VerifiedRunID != rollingRunTwoID {
		t.Fatalf("generation two contract = %#v err=%v", contract, err)
	}
	generation, err := postgresstore.GetReleaseGeneration(ctx, connection, rollingRunTwoID)
	if err != nil || generation.ReleaseID != rollingReleaseTwoID {
		t.Fatalf("active release generation = %#v err=%v", generation, err)
	}
	assertRollingUpgradeTenantPreserved(t, ctx, database)
	ordered, err := postgresstore.FoundationFinalizationOrderVerified(ctx, connection)
	if err != nil || !ordered {
		t.Fatalf("generation two finalization order = %t err=%v", ordered, err)
	}

	restarted := *expand
	recovered, err := restarted.Up(ctx, UpRequest{RunID: rollingRecoveryRunID})
	if err != nil || recovered.RunID != rollingRunTwoID ||
		len(recovered.AppliedVersions) != 0 {
		t.Fatalf("ambiguous commit recovery = %#v err=%v", recovered, err)
	}
	if _, err := postgresstore.GetMigrationRun(
		ctx, connection, rollingRecoveryRunID,
	); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("recovery created a second run: %v", err)
	}
	assertReleaseGenerationConstraints(t, ctx, database)
	assertUnactivatedReleaseRejected(t, ctx, database, expand)
	assertOutOfTransactionReleaseRejected(t, ctx, database, expand)
	assertRollingReleaseForksRejected(t, ctx, database, expand)
}

func configureRollingRelease(
	runner *Runner,
	releaseID string,
	migrationHash string,
	runtimeGeneration int64,
	minimumDatabaseGeneration int64,
	evidence foundationEvidence,
) {
	runner.identity.releaseID = releaseID
	runner.identity.manifestSHA256 = migrationHash
	runner.identity.authoritySHA256 = integrationDigest("authority-" + releaseID)
	runner.evidence = evidence
	runner.schemaAuthority = integrationSchemaAuthority{contract: runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: runtimeGeneration,
		MinimumDatabaseGeneration: minimumDatabaseGeneration,
		MigrationManifestSHA256:   migrationHash,
		ExtensionManifestSHA256:   runner.identity.extensionSHA256,
		PostgresImageDigest:       runtimev1.PostgresImageDigest,
	}}
}

func rollingEvidenceTwo() foundationEvidence {
	return rollingEvidence([]string{
		"018f0000-0000-7000-8000-000000000203",
		"018f0000-0000-7000-8000-000000000204",
		"018f0000-0000-7000-8000-000000000205",
	}, byte('2'))
}

func rollingEvidence(runIDs []string, marker byte) foundationEvidence {
	issuedAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	kinds := []string{"cluster_bootstrap", "extension_admin", "credential_cleanup"}
	receipts := make([]receiptEvidence, 0, len(kinds))
	for index, kind := range kinds {
		document := []byte{'{', '"', marker, byte('a' + index), '"', ':', '1', '}'}
		receipts = append(receipts, receiptEvidence{
			kind: kind, runID: runIDs[index], operationDigest: integrationDigest(kind),
			canonicalDocument: document, canonicalSHA256: integrationBytesDigest(document),
			signatureKeyID: "release_key", issuedAt: issuedAt.Add(time.Duration(index) * time.Second),
		})
	}
	return foundationEvidence{receipts: receipts}
}

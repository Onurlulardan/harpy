package migrator

import (
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) currentReleaseGeneration(
	runID string,
) (postgresstore.ReleaseGenerationRecord, error) {
	var bootstrapRunID, extensionRunID, cleanupRunID string
	for _, receipt := range runner.evidence.receipts {
		switch receipt.kind {
		case "cluster_bootstrap":
			bootstrapRunID = receipt.runID
		case "extension_admin":
			extensionRunID = receipt.runID
		case "credential_cleanup":
			cleanupRunID = receipt.runID
		default:
			return postgresstore.ReleaseGenerationRecord{},
				fmt.Errorf("release generation receipt kind is invalid")
		}
	}
	if bootstrapRunID == "" || extensionRunID == "" || cleanupRunID == "" {
		return postgresstore.ReleaseGenerationRecord{},
			fmt.Errorf("release generation receipt set is incomplete")
	}
	binary := runner.schemaAuthority.Contract()
	identity := runner.identity
	return postgresstore.ReleaseGenerationRecord{
		MigrationRunID: runID, ReleaseID: identity.releaseID,
		EnvironmentID:             identity.environmentID,
		DatabaseOrigin:            string(identity.databaseOrigin),
		DatabaseGeneration:        identity.databaseGeneration,
		MinimumRuntimeGeneration:  identity.minimumRuntimeGeneration,
		RuntimeGeneration:         binary.RuntimeGeneration,
		MinimumDatabaseGeneration: binary.MinimumDatabaseGeneration,
		MigrationSHA256:           identity.manifestSHA256,
		ExtensionSHA256:           identity.extensionSHA256,
		PostgresImageDigest:       binary.PostgresImageDigest,
		BootstrapReceiptRunID:     bootstrapRunID,
		ExtensionReceiptRunID:     extensionRunID,
		CleanupReceiptRunID:       cleanupRunID,
	}, nil
}

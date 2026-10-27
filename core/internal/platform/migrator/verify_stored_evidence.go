package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) verifyStoredEvidence(
	ctx context.Context,
	connection *sql.Conn,
) (postgresstore.CoreSchemaContractRecord, error) {
	environment, err := postgresstore.GetEnvironmentIdentity(ctx, connection)
	if err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	expectedEnvironment := postgresstore.EnvironmentIdentityRecord{
		EnvironmentID:            runner.identity.environmentID,
		DatabaseOrigin:           string(runner.identity.databaseOrigin),
		DatabaseSystemID:         runner.identity.databaseSystemID,
		DatabaseOID:              runner.identity.databaseOID,
		BootstrapReleaseID:       runner.identity.releaseID,
		BootstrapAuthoritySHA256: runner.identity.authoritySHA256,
	}
	if !sameEnvironmentIdentity(environment, expectedEnvironment) {
		return postgresstore.CoreSchemaContractRecord{}, fmt.Errorf("environment evidence differs")
	}
	if err := runner.verifyStoredReceipts(ctx, connection); err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	if err := runner.verifyStoredExtensions(ctx, connection); err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	contract, err := postgresstore.GetCoreSchemaContract(ctx, connection)
	if err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	if err := runner.verifyCoreSchemaContract(ctx, connection, contract); err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	if err := runner.verifyTenantRequirementCatalog(
		ctx, connection, contract.VerifiedRunID,
	); err != nil {
		return postgresstore.CoreSchemaContractRecord{}, err
	}
	return contract, nil
}

func (runner *Runner) verifyStoredReceipts(
	ctx context.Context,
	connection *sql.Conn,
) error {
	for _, receipt := range runner.evidence.receipts {
		actual, err := postgresstore.GetReceiptEvidence(
			ctx, connection, receipt.kind, receipt.runID,
		)
		if err != nil {
			return err
		}
		if !sameReceiptEvidence(actual, runner.receiptRecord(receipt)) {
			return fmt.Errorf("stored %s receipt differs", receipt.kind)
		}
	}
	return nil
}

func (runner *Runner) verifyStoredExtensions(
	ctx context.Context,
	connection *sql.Conn,
) error {
	for _, extension := range runner.evidence.extensions {
		actual, err := postgresstore.GetExtensionEvidence(
			ctx, connection, extension.name, extension.version, extension.verifiedRunID,
		)
		if err != nil {
			return err
		}
		expected := postgresstore.ExtensionEvidenceRecord{
			Name: extension.name, Version: extension.version,
			SchemaName: extension.schemaName, OwnerRole: extension.ownerRole,
			ControlFileSHA256: extension.controlFileSHA256,
			Capabilities:      extension.capabilities,
			DatabaseOrigin:    string(runner.identity.databaseOrigin),
			VerifiedRunID:     extension.verifiedRunID,
		}
		if !sameExtensionEvidence(actual, expected) {
			return fmt.Errorf("stored extension %s differs", extension.name)
		}
	}
	return nil
}

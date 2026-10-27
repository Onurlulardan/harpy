package migrator

import (
	"context"
	"fmt"
	"slices"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistExtensionEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
) error {
	for _, extension := range runner.evidence.extensions {
		expected := postgresstore.ExtensionEvidenceRecord{
			Name: extension.name, Version: extension.version,
			SchemaName: extension.schemaName, OwnerRole: extension.ownerRole,
			ControlFileSHA256: extension.controlFileSHA256,
			Capabilities:      append([]string(nil), extension.capabilities...),
			DatabaseOrigin:    string(runner.identity.databaseOrigin),
			VerifiedRunID:     extension.verifiedRunID,
		}
		if err := postgresstore.CreateExtensionEvidence(ctx, database, expected); err != nil {
			return err
		}
		actual, err := postgresstore.GetExtensionEvidence(
			ctx, database, expected.Name, expected.Version, expected.VerifiedRunID,
		)
		if err != nil {
			return err
		}
		if !sameExtensionEvidence(actual, expected) {
			return fmt.Errorf("stored extension %s differs from verified evidence", expected.Name)
		}
	}
	return nil
}

func sameExtensionEvidence(
	actual, expected postgresstore.ExtensionEvidenceRecord,
) bool {
	return actual.Name == expected.Name && actual.Version == expected.Version &&
		actual.SchemaName == expected.SchemaName && actual.OwnerRole == expected.OwnerRole &&
		actual.ControlFileSHA256 == expected.ControlFileSHA256 &&
		slices.Equal(actual.Capabilities, expected.Capabilities) &&
		actual.DatabaseOrigin == expected.DatabaseOrigin &&
		actual.VerifiedRunID == expected.VerifiedRunID
}

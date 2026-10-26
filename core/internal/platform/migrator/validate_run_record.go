package migrator

import (
	"errors"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) validateRunRecord(record postgresstore.MigrationRunRecord) error {
	identity := runner.identity
	if record.ReleaseID != identity.releaseID ||
		record.EnvironmentID != identity.environmentID ||
		record.DatabaseOrigin != string(identity.databaseOrigin) || record.Command != "up" ||
		record.MigrationManifestSHA256 != identity.manifestSHA256 ||
		record.ExtensionManifestSHA256 != identity.extensionSHA256 ||
		record.RolesManifestSHA256 != identity.rolesSHA256 ||
		record.PrivilegesManifestSHA256 != identity.privilegesSHA256 ||
		record.OwnershipManifestSHA256 != identity.ownershipSHA256 {
		return fmt.Errorf("migration run identity differs from verified release")
	}
	if record.State == "failed" || record.State == "fenced" {
		return errors.Join(
			ErrRunTerminal,
			fmt.Errorf("migration run is %s with code %s", record.State, record.FailureCode),
		)
	}
	return nil
}

func toRunRecord(record postgresstore.MigrationRunRecord) runRecord {
	return runRecord{
		runID:       record.RunID,
		command:     record.Command,
		state:       record.State,
		failureCode: record.FailureCode,
	}
}

package migrator

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const tenantMigrationScope = "tenant"

func validateTenantSelection(selected []postgresv1.SelectedMigration) error {
	if len(selected) == 0 {
		return fmt.Errorf("tenant migration selection is empty")
	}
	for _, candidate := range selected {
		if candidate.Scope != tenantMigrationScope {
			return fmt.Errorf("non-tenant migration entered tenant runner")
		}
		if candidate.TransactionMode != "transactional" {
			return fmt.Errorf("tenant migration %d is not transactional", candidate.Version)
		}
	}
	return nil
}

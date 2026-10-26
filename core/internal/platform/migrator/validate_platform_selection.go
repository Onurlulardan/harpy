package migrator

import (
	"fmt"
	"path"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func validatePlatformSelection(selected []postgresv1.SelectedMigration) error {
	if len(selected) == 0 {
		return fmt.Errorf("platform migration selection is empty")
	}
	lifecycleFound := false
	for _, candidate := range selected {
		if candidate.Scope != platformMigrationScope {
			return fmt.Errorf("non-platform migration entered platform runner")
		}
		if candidate.Version == runLifecycleReadyVersion {
			suffix := "guard_migration_run_transitions." + string(candidate.Origin) + ".sql"
			if candidate.Owner != "platform_migrator" ||
				candidate.TransactionMode != "transactional" ||
				!strings.HasSuffix(path.Base(candidate.Path), suffix) {
				return fmt.Errorf("migration run lifecycle boundary differs from contract")
			}
			lifecycleFound = true
		}
	}
	if !lifecycleFound {
		return fmt.Errorf("migration run lifecycle boundary is absent")
	}
	return nil
}

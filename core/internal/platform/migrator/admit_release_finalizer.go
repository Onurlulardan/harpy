package migrator

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func admitReleaseFinalizer(
	selected []postgresv1.SelectedMigration,
	finalizer ReleaseFinalizer,
) ReleaseFinalizer {
	if finalizer == nil {
		return nil
	}
	for _, migration := range selected {
		if migration.Owner == "catalog" && migration.Scope == platformMigrationScope {
			return finalizer
		}
	}
	return nil
}

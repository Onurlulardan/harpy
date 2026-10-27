package migrator

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func appliedVersionDifference(
	before, after map[int64]historyRow,
	selected []postgresv1.SelectedMigration,
) []int64 {
	versions := make([]int64, 0)
	for _, migration := range selected {
		_, wasApplied := before[migration.Version]
		_, isApplied := after[migration.Version]
		if !wasApplied && isApplied {
			versions = append(versions, migration.Version)
		}
	}
	return versions
}

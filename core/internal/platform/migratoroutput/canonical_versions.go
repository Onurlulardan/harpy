package migratoroutput

import (
	"cmp"
	"fmt"
	"slices"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func canonicalPlanMigrations(
	values []postgresv1.SelectedMigration,
) ([]postgresv1.SelectedMigration, error) {
	byVersion := make(map[int64]postgresv1.SelectedMigration, len(values))
	for _, value := range values {
		if previous, exists := byVersion[value.Version]; exists && previous != value {
			return nil, fmt.Errorf("migration version %d has conflicting output", value.Version)
		}
		byVersion[value.Version] = value
	}
	canonical := make([]postgresv1.SelectedMigration, 0, len(byVersion))
	for _, value := range byVersion {
		canonical = append(canonical, value)
	}
	slices.SortFunc(canonical, func(left, right postgresv1.SelectedMigration) int {
		return cmp.Compare(left.Version, right.Version)
	})
	return canonical, nil
}

func canonicalAppliedVersions(values []int64) []int64 {
	unique := make(map[int64]bool, len(values))
	for _, value := range values {
		unique[value] = true
	}
	canonical := make([]int64, 0, len(unique))
	for value := range unique {
		canonical = append(canonical, value)
	}
	slices.Sort(canonical)
	return canonical
}

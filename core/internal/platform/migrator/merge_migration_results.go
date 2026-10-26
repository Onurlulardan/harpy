package migrator

import (
	"fmt"
	"slices"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func mergeMigrationSelections(
	groups ...[]postgresv1.SelectedMigration,
) ([]postgresv1.SelectedMigration, error) {
	byVersion := map[int64]postgresv1.SelectedMigration{}
	for _, group := range groups {
		for _, selected := range group {
			if previous, exists := byVersion[selected.Version]; exists && previous != selected {
				return nil, fmt.Errorf("migration version %d has conflicting selections", selected.Version)
			}
			byVersion[selected.Version] = selected
		}
	}
	merged := make([]postgresv1.SelectedMigration, 0, len(byVersion))
	for _, selected := range byVersion {
		merged = append(merged, selected)
	}
	slices.SortFunc(merged, func(left, right postgresv1.SelectedMigration) int {
		switch {
		case left.Version < right.Version:
			return -1
		case left.Version > right.Version:
			return 1
		default:
			return 0
		}
	})
	return merged, nil
}

func mergeAppliedVersions(groups ...[]int64) []int64 {
	unique := map[int64]bool{}
	for _, group := range groups {
		for _, version := range group {
			unique[version] = true
		}
	}
	merged := make([]int64, 0, len(unique))
	for version := range unique {
		merged = append(merged, version)
	}
	slices.Sort(merged)
	return merged
}

package postgresv1

import "fmt"

func (manifest MigrationManifest) Select(origin DatabaseOrigin) ([]SelectedMigration, error) {
	if !origin.Valid() {
		return nil, fmt.Errorf("database origin is invalid")
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	selected := make([]SelectedMigration, 0, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		variant := entry.Variants.Fresh
		if origin == DatabaseOriginV1Adopted {
			variant = entry.Variants.V1Adopted
		}
		if variant == nil {
			return nil, fmt.Errorf("migration %d has no %s variant", entry.Version, origin)
		}
		selected = append(selected, SelectedMigration{
			Owner: entry.Owner, Scope: entry.Scope, Version: entry.Version,
			TransactionMode: entry.TransactionMode, Phase: entry.Phase,
			TargetDatabaseGeneration:     entry.TargetDatabaseGeneration,
			NextMinimumRuntimeGeneration: entry.NextMinimumRuntimeGeneration,
			Origin:                       origin, Path: variant.Path, SHA256: variant.SHA256,
		})
	}
	return selected, nil
}

package migrator

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type releaseGenerationBoundary struct {
	databaseGeneration       int64
	minimumRuntimeGeneration int64
}

func selectReleaseGenerationBoundary(
	platform []postgresv1.SelectedMigration,
	tenant []postgresv1.SelectedMigration,
) (releaseGenerationBoundary, error) {
	if len(platform) == 0 {
		return releaseGenerationBoundary{}, fmt.Errorf(
			"platform release migration generation boundary is absent",
		)
	}
	var boundary releaseGenerationBoundary
	for _, migration := range platform {
		if migration.TargetDatabaseGeneration > boundary.databaseGeneration {
			boundary.databaseGeneration = migration.TargetDatabaseGeneration
		}
		if migration.NextMinimumRuntimeGeneration > boundary.minimumRuntimeGeneration {
			boundary.minimumRuntimeGeneration = migration.NextMinimumRuntimeGeneration
		}
	}
	if boundary.databaseGeneration < 1 || boundary.minimumRuntimeGeneration < 1 {
		return releaseGenerationBoundary{}, fmt.Errorf(
			"release migration generation boundary is absent",
		)
	}
	for _, migration := range tenant {
		if migration.TargetDatabaseGeneration > boundary.databaseGeneration ||
			migration.NextMinimumRuntimeGeneration > boundary.minimumRuntimeGeneration {
			return releaseGenerationBoundary{}, fmt.Errorf(
				"tenant migration exceeds the platform release generation boundary",
			)
		}
	}
	return boundary, nil
}

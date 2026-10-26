package postgres

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func validateMigrationMetadata(
	runID string,
	selected []postgresv1.SelectedMigration,
) (map[int64]migrationMetadata, error) {
	if !validCanonicalRunID(runID) {
		return nil, fmt.Errorf("migration run ID is invalid")
	}
	canonical, err := canonicalSelectedMigrations(selected)
	if err != nil {
		return nil, err
	}
	metadata := make(map[int64]migrationMetadata, len(canonical))
	for _, migration := range canonical {
		metadata[migration.Version] = migrationMetadata{
			owner: migration.Owner, scope: migration.Scope, version: migration.Version,
			origin: migration.Origin, selectedPath: migration.Path,
			selectedSHA256: migration.SHA256, runID: runID,
			targetDatabaseGeneration:     migration.TargetDatabaseGeneration,
			nextMinimumRuntimeGeneration: migration.NextMinimumRuntimeGeneration,
		}
	}
	return metadata, nil
}

func validCanonicalRunID(value string) bool {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value ||
		parsed.Variant() != uuid.RFC4122 {
		return false
	}
	version := parsed.Version()
	return version >= 1 && version <= 8
}

func canonicalSelectedMigrations(
	selected []postgresv1.SelectedMigration,
) ([]postgresv1.SelectedMigration, error) {
	if len(selected) == 0 {
		return nil, fmt.Errorf("selected migrations are empty")
	}
	origin := selected[0].Origin
	entries := make([]postgresv1.MigrationEntry, 0, len(selected))
	for _, migration := range selected {
		if migration.Origin != origin {
			return nil, fmt.Errorf("selected migration origins differ")
		}
		variant := &postgresv1.MigrationVariant{
			Path: migration.Path, SHA256: migration.SHA256,
		}
		variants := postgresv1.MigrationVariants{Fresh: variant}
		if origin == postgresv1.DatabaseOriginV1Adopted {
			variants = postgresv1.MigrationVariants{V1Adopted: variant}
		}
		entries = append(entries, postgresv1.MigrationEntry{
			Owner: migration.Owner, Scope: migration.Scope, Version: migration.Version,
			TransactionMode: migration.TransactionMode, Phase: migration.Phase,
			TargetDatabaseGeneration:     migration.TargetDatabaseGeneration,
			NextMinimumRuntimeGeneration: migration.NextMinimumRuntimeGeneration,
			Variants:                     variants,
		})
	}
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: historyTableName, OutOfOrder: false,
		Entries: entries,
	}
	canonical, err := manifest.Select(origin)
	if err != nil {
		return nil, fmt.Errorf("validate selected migrations: %w", err)
	}
	if !slices.Equal(canonical, selected) {
		return nil, fmt.Errorf("selected migrations are not canonical")
	}
	return canonical, nil
}

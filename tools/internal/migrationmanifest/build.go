package migrationmanifest

import (
	"fmt"
	"sort"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type logicalMigration struct {
	entry           postgresv1.MigrationEntry
	logicalName     string
	transactionMode string
}

func Build(
	inventory StorageInventory,
	groups []SourceGroup,
) (postgresv1.MigrationManifest, error) {
	if len(groups) == 0 {
		return postgresv1.MigrationManifest{}, fmt.Errorf("migration source groups are empty")
	}
	if err := validateSourceClaims(inventory, groups); err != nil {
		return postgresv1.MigrationManifest{}, err
	}
	logical := make(map[int64]*logicalMigration)
	for groupIndex, group := range groups {
		if len(group.Documents) == 0 {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"migration source group %d is empty", groupIndex,
			)
		}
		for _, document := range group.Documents {
			if err := addDocument(logical, group, document); err != nil {
				return postgresv1.MigrationManifest{}, err
			}
		}
	}
	versions := make([]int64, 0, len(logical))
	for version := range logical {
		versions = append(versions, version)
	}
	sort.Slice(versions, func(left, right int) bool { return versions[left] < versions[right] })
	manifest := postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1",
		Engine:          "goose.v3",
		PostgresMajor:   18,
		HistoryTable:    "platform.platform_migration_history",
		Entries:         make([]postgresv1.MigrationEntry, 0, len(versions)),
	}
	for _, version := range versions {
		candidate := logical[version]
		if candidate.entry.Variants.Fresh == nil || candidate.entry.Variants.V1Adopted == nil {
			return postgresv1.MigrationManifest{}, fmt.Errorf(
				"migration %d does not have both origin variants", version,
			)
		}
		manifest.Entries = append(manifest.Entries, candidate.entry)
	}
	if err := manifest.Validate(); err != nil {
		return postgresv1.MigrationManifest{}, fmt.Errorf("validate generated manifest: %w", err)
	}
	return manifest, nil
}

func addDocument(
	logical map[int64]*logicalMigration,
	group SourceGroup,
	document SourceDocument,
) error {
	if document.Owner() != group.Owner {
		return fmt.Errorf("migration %s owner differs from source group", document.Path())
	}
	identity, err := parseDocumentIdentity(document.Path())
	if err != nil {
		return fmt.Errorf("migration %s: %w", document.Path(), err)
	}
	metadata, err := postgresv1.InspectMigrationSource(document.Content())
	if err != nil {
		return fmt.Errorf("migration %s: %w", document.Path(), err)
	}
	candidate, exists := logical[identity.version]
	if !exists {
		candidate = newLogicalMigration(group, identity, metadata.TransactionMode)
		logical[identity.version] = candidate
	} else if candidate.entry.Owner != group.Owner ||
		candidate.entry.Scope != group.Scope ||
		candidate.entry.Phase != group.Phase ||
		candidate.entry.TargetDatabaseGeneration != group.TargetDatabaseGeneration ||
		candidate.entry.NextMinimumRuntimeGeneration != group.NextMinimumRuntimeGeneration ||
		candidate.logicalName != identity.logicalName ||
		candidate.transactionMode != metadata.TransactionMode {
		return fmt.Errorf("migration version %d has conflicting identity", identity.version)
	}
	variant := &postgresv1.MigrationVariant{Path: document.Path(), SHA256: metadata.SHA256}
	if identity.origin == postgresv1.DatabaseOriginFresh {
		if candidate.entry.Variants.Fresh != nil {
			return fmt.Errorf("migration version %d duplicates fresh variant", identity.version)
		}
		candidate.entry.Variants.Fresh = variant
	} else {
		if candidate.entry.Variants.V1Adopted != nil {
			return fmt.Errorf("migration version %d duplicates adopted variant", identity.version)
		}
		candidate.entry.Variants.V1Adopted = variant
	}
	return nil
}

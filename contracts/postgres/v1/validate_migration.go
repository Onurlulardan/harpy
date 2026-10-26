package postgresv1

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (manifest MigrationManifest) Validate() error {
	if manifest.ContractVersion != "postgres-migration-manifest.v1" ||
		manifest.Engine != "goose.v3" || manifest.PostgresMajor != 18 ||
		manifest.HistoryTable != "platform.platform_migration_history" || manifest.OutOfOrder {
		return fmt.Errorf("migration manifest header is invalid")
	}
	if len(manifest.Entries) == 0 || len(manifest.Entries) > 10000 {
		return fmt.Errorf("migration entries are invalid")
	}
	paths := make(map[string]struct{}, len(manifest.Entries)*2)
	var previousVersion, previousDatabase, previousRuntime int64
	for index, entry := range manifest.Entries {
		if err := validateMigrationEntry(entry, paths); err != nil {
			return fmt.Errorf("migration entry %d: %w", index, err)
		}
		if index > 0 && entry.Version <= previousVersion {
			return fmt.Errorf("migration entries are not strictly version ordered")
		}
		if entry.TargetDatabaseGeneration < previousDatabase ||
			entry.NextMinimumRuntimeGeneration < previousRuntime {
			return fmt.Errorf("migration generations regress")
		}
		previousVersion = entry.Version
		previousDatabase = entry.TargetDatabaseGeneration
		previousRuntime = entry.NextMinimumRuntimeGeneration
	}
	return nil
}

func validateMigrationEntry(entry MigrationEntry, paths map[string]struct{}) error {
	if err := validateModule(entry.Owner, "owner"); err != nil {
		return err
	}
	if entry.Scope != "platform" && entry.Scope != "tenant" {
		return fmt.Errorf("scope is invalid")
	}
	if entry.Version < 20000101000000 || entry.Version > 99991231235959 {
		return fmt.Errorf("version is invalid")
	}
	if _, err := time.Parse("20060102150405", strconv.FormatInt(entry.Version, 10)); err != nil {
		return fmt.Errorf("version is not a UTC timestamp")
	}
	if entry.TransactionMode != "transactional" && entry.TransactionMode != "non_transactional" {
		return fmt.Errorf("transaction_mode is invalid")
	}
	if !validMigrationPhase(entry.Phase) || entry.TargetDatabaseGeneration < 1 ||
		entry.TargetDatabaseGeneration > maximumSafeInteger ||
		entry.NextMinimumRuntimeGeneration < 1 ||
		entry.NextMinimumRuntimeGeneration > maximumSafeInteger {
		return fmt.Errorf("phase or generation is invalid")
	}
	if entry.Variants.Fresh == nil && entry.Variants.V1Adopted == nil {
		return fmt.Errorf("at least one origin variant is required")
	}
	variants := []struct {
		origin  DatabaseOrigin
		variant *MigrationVariant
	}{
		{DatabaseOriginFresh, entry.Variants.Fresh},
		{DatabaseOriginV1Adopted, entry.Variants.V1Adopted},
	}
	for _, candidate := range variants {
		if candidate.variant == nil {
			continue
		}
		if err := validateMigrationVariant(entry.Version, candidate.origin, *candidate.variant); err != nil {
			return err
		}
		if _, exists := paths[candidate.variant.Path]; exists {
			return fmt.Errorf("migration path is duplicated")
		}
		paths[candidate.variant.Path] = struct{}{}
	}
	return nil
}

func validateMigrationVariant(version int64, origin DatabaseOrigin, variant MigrationVariant) error {
	if err := validateArtifactPath(variant.Path, "variant path"); err != nil {
		return err
	}
	base := variant.Path[strings.LastIndex(variant.Path, "/")+1:]
	prefix := strconv.FormatInt(version, 10) + "_"
	if !strings.HasPrefix(base, prefix) || !strings.HasSuffix(base, "."+string(origin)+".sql") {
		return fmt.Errorf("variant path does not bind version and origin")
	}
	return validateDigest(variant.SHA256, "variant sha256")
}

func validMigrationPhase(value string) bool {
	switch value {
	case "foundation", "expand", "backfill", "switch", "contract":
		return true
	default:
		return false
	}
}

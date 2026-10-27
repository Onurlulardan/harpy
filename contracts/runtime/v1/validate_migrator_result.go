package runtimev1

import "fmt"

func validatePlanResult(result MigratorCommandResult) error {
	if result.Plan == nil || result.Verify != nil || result.Up != nil ||
		result.Plan.CurrentVersion < 0 || result.Plan.Pending == nil {
		return fmt.Errorf("migration plan result is invalid")
	}
	previousVersion := int64(0)
	for _, entry := range result.Plan.Pending {
		if !migratorOwnerPattern.MatchString(entry.Owner) ||
			(entry.Scope != "platform" && entry.Scope != "tenant") ||
			entry.Version <= previousVersion ||
			(entry.TransactionMode != "transactional" && entry.TransactionMode != "non_transactional") ||
			entry.Phase == "" || len(entry.Phase) > 64 || entry.TargetDatabaseGeneration < 1 ||
			entry.NextMinimumRuntimeGeneration < 1 ||
			(entry.Origin != "fresh" && entry.Origin != "v1_adopted") ||
			len(entry.Path) > 512 || !migratorPathPattern.MatchString(entry.Path) ||
			!validMigratorDigest(entry.SHA256) {
			return fmt.Errorf("migration plan entry is invalid")
		}
		previousVersion = entry.Version
	}
	return nil
}

func validateVerifyResult(result MigratorCommandResult) error {
	value := result.Verify
	if value == nil || result.Plan != nil || result.Up != nil ||
		value.DatabaseVersion < 1 || value.DatabaseGeneration < 1 ||
		value.MinimumRuntimeGeneration < 1 || !validMigratorUUID(value.VerifiedRunID) {
		return fmt.Errorf("migration verify result is invalid")
	}
	return nil
}

func validateUpResult(result MigratorCommandResult) error {
	value := result.Up
	if value == nil || result.Plan != nil || result.Verify != nil ||
		!validMigratorUUID(value.RunID) || value.AppliedVersions == nil ||
		value.DatabaseVersion < 1 {
		return fmt.Errorf("migration up result is invalid")
	}
	previous := int64(0)
	for _, version := range value.AppliedVersions {
		if version <= previous {
			return fmt.Errorf("migration up applied versions are invalid")
		}
		previous = version
	}
	return nil
}

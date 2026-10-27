package postgres

import (
	"context"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type foundationPreflight struct {
	count                    int64
	origin                   string
	systemIdentifier         string
	databaseOID              int64
	environmentID            string
	bootstrapReleaseID       string
	activeReleaseID          string
	runID                    string
	runState                 string
	runCommand               string
	runMigrationSHA256       string
	runExtensionSHA256       string
	databaseGeneration       int64
	minimumRuntimeGeneration int64
	contractMigrationSHA256  string
	contractExtensionSHA256  string
	contractStatus           string
	contractVerifiedRunID    string
	contractRevision         int64
	extensionVersion         string
	extensionSchema          string
	extensionOwner           string
	extensionOrigin          string
	extensionVerifiedRunID   string
	binaryReleaseKnown       bool
}

func validateFoundationPreflight(
	ctx context.Context,
	queryer roleQueryer,
	catalog catalogPreflight,
	binary runtimev1.BinarySchemaContract,
) (runtimev1.DatabaseOrigin, error) {
	result, err := readFoundationPreflight(ctx, queryer, binary)
	if err != nil {
		return "", err
	}
	origin, err := validateFoundationEvidence(result, catalog)
	if err != nil {
		return "", err
	}
	if reason := foundationSchemaReadiness(result, binary); reason != SchemaReady {
		return "", fmt.Errorf("%w: schema contract is %s", ErrStartupPreflight, reason)
	}
	return origin, nil
}

func validateFoundationEvidence(
	result foundationPreflight,
	catalog catalogPreflight,
) (runtimev1.DatabaseOrigin, error) {
	origin := runtimev1.DatabaseOrigin(result.origin)
	if result.count != 1 || !origin.Valid() || result.runState != "succeeded" ||
		result.runCommand != "up" ||
		result.systemIdentifier != catalog.systemIdentifier ||
		result.databaseOID != catalog.databaseOID ||
		result.runID != result.contractVerifiedRunID ||
		result.runMigrationSHA256 != result.contractMigrationSHA256 ||
		result.runExtensionSHA256 != result.contractExtensionSHA256 ||
		result.extensionVersion != catalog.vectorVersion ||
		result.extensionSchema != catalog.vectorSchema ||
		result.extensionOwner != catalog.vectorOwner ||
		result.extensionOrigin != result.origin ||
		result.extensionVerifiedRunID == "" || result.environmentID == "" ||
		result.bootstrapReleaseID == "" || result.activeReleaseID == "" {
		return "", fmt.Errorf("%w: foundation evidence contract", ErrStartupPreflight)
	}
	if err := validateExtensionPlacement(origin, result.extensionSchema, result.extensionOwner); err != nil {
		return "", err
	}
	return origin, nil
}

func foundationSchemaReadiness(
	result foundationPreflight,
	binary runtimev1.BinarySchemaContract,
) SchemaReadinessReason {
	return evaluateSchemaContract(
		binary, result.databaseGeneration, result.minimumRuntimeGeneration,
		result.contractMigrationSHA256, result.contractExtensionSHA256,
		result.contractStatus, result.contractVerifiedRunID, result.contractRevision,
		result.binaryReleaseKnown,
	)
}

func readFoundationPreflight(
	ctx context.Context,
	queryer roleQueryer,
	binary runtimev1.BinarySchemaContract,
) (foundationPreflight, error) {
	var result foundationPreflight
	err := queryer.QueryRow(
		ctx, preflightFoundationSQL, runtimev1.VectorExtensionVersion,
		binary.RuntimeGeneration, binary.MinimumDatabaseGeneration,
		binary.MigrationManifestSHA256, binary.ExtensionManifestSHA256,
		binary.PostgresImageDigest,
	).Scan(
		&result.count, &result.origin, &result.systemIdentifier, &result.databaseOID,
		&result.environmentID, &result.bootstrapReleaseID, &result.activeReleaseID,
		&result.runID, &result.runState,
		&result.runCommand,
		&result.runMigrationSHA256, &result.runExtensionSHA256,
		&result.databaseGeneration, &result.minimumRuntimeGeneration,
		&result.contractMigrationSHA256, &result.contractExtensionSHA256,
		&result.contractStatus, &result.contractVerifiedRunID, &result.contractRevision,
		&result.extensionVersion, &result.extensionSchema, &result.extensionOwner,
		&result.extensionOrigin, &result.extensionVerifiedRunID,
		&result.binaryReleaseKnown,
	)
	if err != nil {
		return foundationPreflight{}, fmt.Errorf("%w: foundation evidence query", ErrStartupPreflight)
	}
	return result, nil
}

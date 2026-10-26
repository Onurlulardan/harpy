package postgres

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (pool *Pool) ProbeSchemaContract(
	ctx context.Context,
) (SchemaReadiness, error) {
	if pool == nil || pool.closed.Load() {
		return SchemaReadiness{}, ErrPoolClosed
	}
	return probeSchemaContractRow(
		pool.inner.QueryRow(
			ctx, probeSchemaContractSQL,
			pool.binaryContract.RuntimeGeneration,
			pool.binaryContract.MinimumDatabaseGeneration,
			pool.binaryContract.MigrationManifestSHA256,
			pool.binaryContract.ExtensionManifestSHA256,
			pool.binaryContract.PostgresImageDigest,
		), pool.binaryContract,
	)
}

type schemaContractRow interface {
	Scan(destinations ...any) error
}

func probeSchemaContractRow(
	row schemaContractRow,
	contract runtimev1.BinarySchemaContract,
) (SchemaReadiness, error) {
	var databaseGeneration, minimumRuntimeGeneration, revision int64
	var migrationHash, extensionHash, status, verifiedRunID string
	var binaryReleaseKnown bool
	err := row.Scan(
		&databaseGeneration, &minimumRuntimeGeneration,
		&migrationHash, &extensionHash, &status, &verifiedRunID, &revision,
		&binaryReleaseKnown,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SchemaReadiness{reason: SchemaMissing}, nil
	}
	if err != nil {
		return SchemaReadiness{}, fmt.Errorf("schema contract probe failed")
	}
	result := SchemaReadiness{
		reason: SchemaReady, databaseGeneration: databaseGeneration,
		revision: revision, verifiedRunID: verifiedRunID,
	}
	result.reason = evaluateSchemaContract(
		contract, databaseGeneration, minimumRuntimeGeneration,
		migrationHash, extensionHash, status, verifiedRunID, revision,
		binaryReleaseKnown,
	)
	result.ready = result.reason == SchemaReady
	return result, nil
}

func evaluateSchemaContract(
	contract runtimev1.BinarySchemaContract,
	databaseGeneration int64,
	minimumRuntimeGeneration int64,
	migrationHash string,
	extensionHash string,
	status string,
	verifiedRunID string,
	revision int64,
	binaryReleaseKnown bool,
) SchemaReadinessReason {
	if databaseGeneration <= 0 || minimumRuntimeGeneration <= 0 || revision <= 0 ||
		uuid.Validate(verifiedRunID) != nil || !sha256Pattern.MatchString(migrationHash) ||
		!sha256Pattern.MatchString(extensionHash) {
		return SchemaContractMalformed
	}
	if status != "ready" {
		return SchemaFenced
	}
	if databaseGeneration < contract.MinimumDatabaseGeneration {
		return SchemaDatabaseTooOld
	}
	if contract.RuntimeGeneration < minimumRuntimeGeneration {
		return SchemaRuntimeTooOld
	}
	if !binaryReleaseKnown {
		return SchemaManifestMismatch
	}
	return SchemaReady
}

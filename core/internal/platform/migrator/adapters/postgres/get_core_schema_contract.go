package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_core_schema_contract.sql
var getCoreSchemaContractSQL string

func GetCoreSchemaContract(
	ctx context.Context,
	database goosedb.DBTxConn,
) (CoreSchemaContractRecord, error) {
	var record CoreSchemaContractRecord
	if err := database.QueryRowContext(ctx, getCoreSchemaContractSQL).Scan(
		&record.DatabaseGeneration, &record.MinimumRuntimeGeneration,
		&record.MigrationSHA256, &record.ExtensionSHA256,
		&record.Status, &record.VerifiedRunID, &record.Revision,
	); err != nil {
		return CoreSchemaContractRecord{}, fmt.Errorf("get core schema contract: %w", err)
	}
	return record, nil
}

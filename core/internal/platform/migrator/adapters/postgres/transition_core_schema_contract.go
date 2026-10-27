package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/transition_core_schema_contract.sql
var transitionCoreSchemaContractSQL string

func TransitionCoreSchemaContract(
	ctx context.Context,
	database goosedb.DBTxConn,
	record CoreSchemaContractRecord,
) (CoreSchemaContractRecord, error) {
	var stored CoreSchemaContractRecord
	err := database.QueryRowContext(
		ctx, transitionCoreSchemaContractSQL,
		record.DatabaseGeneration, record.MinimumRuntimeGeneration,
		record.MigrationSHA256, record.ExtensionSHA256,
		record.VerifiedRunID, record.Status,
	).Scan(
		&stored.DatabaseGeneration, &stored.MinimumRuntimeGeneration,
		&stored.MigrationSHA256, &stored.ExtensionSHA256,
		&stored.Status, &stored.VerifiedRunID, &stored.Revision,
	)
	if err != nil {
		return CoreSchemaContractRecord{},
			fmt.Errorf("transition core schema contract: %w", err)
	}
	return stored, nil
}

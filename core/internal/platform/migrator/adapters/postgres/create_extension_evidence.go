package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_extension_evidence.sql
var createExtensionEvidenceSQL string

func CreateExtensionEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
	record ExtensionEvidenceRecord,
) error {
	result, err := database.ExecContext(
		ctx, createExtensionEvidenceSQL,
		record.Name, record.Version, record.SchemaName, record.OwnerRole,
		record.ControlFileSHA256, record.Capabilities, record.DatabaseOrigin,
		record.VerifiedRunID,
	)
	if err != nil {
		return fmt.Errorf("create extension %s evidence: %w", record.Name, err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create extension evidence returned an invalid row count")
	}
	return nil
}

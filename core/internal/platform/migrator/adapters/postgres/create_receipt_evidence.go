package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/create_receipt_evidence.sql
var createReceiptEvidenceSQL string

func CreateReceiptEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
	record ReceiptEvidenceRecord,
) error {
	result, err := database.ExecContext(
		ctx, createReceiptEvidenceSQL,
		record.Kind, record.RunID, record.EnvironmentID, record.ReleaseID,
		record.DatabaseOrigin, record.DatabaseSystemID, record.DatabaseOID,
		record.OperationDigest, record.CanonicalDocument, record.CanonicalSHA256,
		record.SignatureKeyID, record.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("create %s receipt evidence: %w", record.Kind, err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected > 1 {
		return fmt.Errorf("create receipt evidence returned an invalid row count")
	}
	return nil
}

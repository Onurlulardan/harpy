package postgres

import (
	"context"
	_ "embed"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_receipt_evidence.sql
var getReceiptEvidenceSQL string

func GetReceiptEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
	kind, runID string,
) (ReceiptEvidenceRecord, error) {
	record := ReceiptEvidenceRecord{Kind: kind, RunID: runID}
	if err := database.QueryRowContext(ctx, getReceiptEvidenceSQL, kind, runID).Scan(
		&record.EnvironmentID, &record.ReleaseID, &record.DatabaseOrigin,
		&record.DatabaseSystemID, &record.DatabaseOID, &record.OperationDigest,
		&record.CanonicalDocument, &record.CanonicalSHA256,
		&record.SignatureKeyID, &record.IssuedAt,
	); err != nil {
		return ReceiptEvidenceRecord{}, fmt.Errorf("get %s receipt evidence: %w", kind, err)
	}
	return record, nil
}

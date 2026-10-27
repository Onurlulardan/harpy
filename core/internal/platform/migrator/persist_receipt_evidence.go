package migrator

import (
	"bytes"
	"context"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	goosedb "github.com/pressly/goose/v3/database"
)

func (runner *Runner) persistReceiptEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
) error {
	for _, receipt := range runner.evidence.receipts {
		expected := runner.receiptRecord(receipt)
		if err := postgresstore.CreateReceiptEvidence(ctx, database, expected); err != nil {
			return err
		}
		actual, err := postgresstore.GetReceiptEvidence(
			ctx, database, receipt.kind, receipt.runID,
		)
		if err != nil {
			return err
		}
		if !sameReceiptEvidence(actual, expected) {
			return fmt.Errorf("stored %s receipt differs from verified evidence", receipt.kind)
		}
	}
	return nil
}

func (runner *Runner) receiptRecord(
	receipt receiptEvidence,
) postgresstore.ReceiptEvidenceRecord {
	return postgresstore.ReceiptEvidenceRecord{
		Kind: receipt.kind, RunID: receipt.runID,
		EnvironmentID:     runner.identity.environmentID,
		ReleaseID:         runner.identity.releaseID,
		DatabaseOrigin:    string(runner.identity.databaseOrigin),
		DatabaseSystemID:  runner.identity.databaseSystemID,
		DatabaseOID:       runner.identity.databaseOID,
		OperationDigest:   receipt.operationDigest,
		CanonicalDocument: append([]byte(nil), receipt.canonicalDocument...),
		CanonicalSHA256:   receipt.canonicalSHA256,
		SignatureKeyID:    receipt.signatureKeyID, IssuedAt: receipt.issuedAt,
	}
}

func sameReceiptEvidence(
	actual, expected postgresstore.ReceiptEvidenceRecord,
) bool {
	return actual.Kind == expected.Kind && actual.RunID == expected.RunID &&
		actual.EnvironmentID == expected.EnvironmentID && actual.ReleaseID == expected.ReleaseID &&
		actual.DatabaseOrigin == expected.DatabaseOrigin &&
		actual.DatabaseSystemID == expected.DatabaseSystemID &&
		actual.DatabaseOID == expected.DatabaseOID &&
		actual.OperationDigest == expected.OperationDigest &&
		bytes.Equal(actual.CanonicalDocument, expected.CanonicalDocument) &&
		actual.CanonicalSHA256 == expected.CanonicalSHA256 &&
		actual.SignatureKeyID == expected.SignatureKeyID &&
		actual.IssuedAt.Equal(expected.IssuedAt)
}

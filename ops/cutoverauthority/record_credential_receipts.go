package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const recordCredentialReceiptsSQL = `SELECT
cutover_authority.record_credential_receipts_v1(
 $1::uuid,$2::uuid,$3::text,$4::bytea,$5::text,$6::bytea)`

func recordCredentialReceipts(
	ctx context.Context, transaction pgx.Tx,
	prerequisites verifiedFinalizationPrerequisites,
) error {
	_, err := transaction.Exec(ctx, recordCredentialReceiptsSQL,
		prerequisites.adoption.OperationID(), prerequisites.adoption.RunID(),
		prerequisites.seal.ReceiptSHA256(), prerequisites.documents.seal,
		prerequisites.cleanup.ReceiptSHA256(), prerequisites.documents.cleanup,
	)
	if err != nil {
		return fmt.Errorf("credential receipt persistence failed")
	}
	return nil
}

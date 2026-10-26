package tenantmigrationexecutor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func (executor *Executor) recordFailure(
	ctx context.Context, claim Claim, cause error,
) error {
	outcome, code := failureDetails(cause)
	receiptID, err := executor.newID()
	if err != nil {
		return err
	}
	evidence := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%d:%s:%s", claim.JobID, claim.Attempt, claim.MigrationRunID, code,
	)))
	digest := hex.EncodeToString(evidence[:])
	return executor.store.Fail(ctx, claim, Failure{
		Outcome: outcome, Code: code, ReceiptID: receiptID,
		ReceiptSHA256: digest, EvidenceSHA256: digest,
	})
}

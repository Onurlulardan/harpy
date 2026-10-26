package billingv1adoptionapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func RehearseRollback(
	ctx context.Context,
	store ScopeStore,
	operation Operation,
	receiptID uuid.UUID,
	observedSourceSHA256 string,
	rehearsedAt time.Time,
) (RollbackRehearsal, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !validOperation(operation) ||
		receiptID == uuid.Nil || !validDigest(observedSourceSHA256) || !canonicalInstant(rehearsedAt) {
		return RollbackRehearsal{}, fmt.Errorf("Billing V1 rollback rehearsal is not configured")
	}
	result, err := store.RehearseRollback(
		ctx, operation, receiptID, observedSourceSHA256, rehearsedAt,
	)
	if err != nil {
		return RollbackRehearsal{}, fmt.Errorf("rehearse Billing V1 rollback: %w", err)
	}
	if result.ReceiptID != receiptID || result.RehearsedAt != rehearsedAt ||
		result.SourceSHA256 != observedSourceSHA256 || result.ReverseSHA256 != result.SourceSHA256 {
		return RollbackRehearsal{}, fmt.Errorf("Billing V1 rollback rehearsal is not zero-diff")
	}
	return result, nil
}

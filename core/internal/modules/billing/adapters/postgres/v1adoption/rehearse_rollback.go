package v1adoption

import (
	"context"
	"fmt"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

func (storage *store) RehearseRollback(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
	receiptID uuid.UUID,
	observedSourceSHA256 string,
	rehearsedAt time.Time,
) (billingv1adoptionapp.RollbackRehearsal, error) {
	var result billingv1adoptionapp.RollbackRehearsal
	err := storage.database.QueryRow(
		ctx, rehearseRollbackSQL, operation.OperationID, operation.RunID,
		receiptID, observedSourceSHA256, rehearsedAt,
	).Scan(
		&result.ReceiptID, &result.SourceSHA256,
		&result.ReverseSHA256, &result.RehearsedAt,
	)
	if err != nil {
		return billingv1adoptionapp.RollbackRehearsal{},
			fmt.Errorf("rehearse Billing V1 adoption rollback: %w", err)
	}
	result.RehearsedAt = result.RehearsedAt.UTC()
	return result, nil
}

package lifecycle

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func recordPurgeOwnerResult(
	ctx context.Context, store Store, operation catalogdomain.Operation,
	pin catalogdomain.ComponentPin, result componentowner.OwnerResult,
) error {
	return RecordOwnerReceipt(ctx, store, catalogdomain.OwnerReceipt{
		ReceiptID: mustStableID(result.ReceiptID), OperationID: operation.ID,
		ComponentID: pin.ComponentID, Owner: pin.Owner,
		OwnerGeneration: pin.OwnerGeneration, Contract: pin.Contract,
		ContentSHA256: pin.ContentSHA256, Action: "purge", Ready: true,
		CompletedAt: result.CompletedAt,
	})
}

func earlierRetry(current, candidate time.Time) time.Time {
	if current.IsZero() || candidate.Before(current) {
		return candidate
	}
	return current
}

package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) providerReceiptApplied(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.ProviderExecution,
	plan subscriptionapp.CommitPlan,
) error {
	err := storage.expectID(
		ctx, completeProviderReceiptSQL, execution.ProviderReceiptID,
		authorityID, storage.tenantID, execution.ProviderReceiptID,
		execution.NormalizedFactID, plan.Receipt.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("complete billing provider receipt: %w", err)
	}
	return nil
}

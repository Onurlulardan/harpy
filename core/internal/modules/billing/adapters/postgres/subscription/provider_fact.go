package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) assertProviderFact(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.ProviderExecution,
	replay bool,
) error {
	var factID uuid.UUID
	var processorState string
	err := storage.queryRow(
		ctx, assertProviderFactSQL, authorityID, storage.tenantID,
		execution.NormalizedFactID, execution.ProviderReceiptID,
		execution.NormalizedFactSHA[:], execution.SubscriptionID,
	).Scan(&factID, &processorState)
	if err != nil {
		return fmt.Errorf("assert billing provider fact: %w", err)
	}
	expectedState := "processing"
	if replay {
		expectedState = "applied"
	}
	if factID != execution.NormalizedFactID || processorState != expectedState {
		return subscriptionapp.ErrProviderFactReused
	}
	return nil
}

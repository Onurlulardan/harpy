package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (storage *postgresStore) RecordReceipt(
	ctx context.Context,
	receipt tenantapi.OnboardingReceipt,
	expectedRevision int64,
) (tenantapi.OnboardingTransition, error) {
	if !storage.usable(ctx) || !receipt.IsValid() || expectedRevision < 1 {
		return tenantapi.OnboardingTransition{}, errUnavailable
	}
	transition, err := scanOnboardingTransition(storage.queryRow(ctx, recordSQL,
		uuid.UUID(receipt.ReceiptID), uuid.UUID(receipt.OperationID),
		uuid.UUID(receipt.AttemptID), string(receipt.Kind),
		receipt.SHA256, receipt.EvidenceSHA256, receipt.OccurredAt, expectedRevision))
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("record onboarding row: %w", err)
	}
	return transition, nil
}

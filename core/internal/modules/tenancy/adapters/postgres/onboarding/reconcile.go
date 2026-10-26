package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (storage *postgresStore) Reconcile(
	ctx context.Context,
	request onboardingapp.ReconcileRequest,
) (tenantapi.OnboardingTransition, error) {
	if !storage.usable(ctx) {
		return tenantapi.OnboardingTransition{}, errUnavailable
	}
	transition, err := scanOnboardingTransition(storage.queryRow(ctx, reconcileSQL,
		request.OperationID, request.ReceiptID, request.ReceiptSHA256,
		request.EvidenceSHA256, request.ExpectedRevision))
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("reconcile onboarding row: %w", err)
	}
	return transition, nil
}

package onboarding

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) Reconcile(
	ctx context.Context,
	request ReconcileRequest,
) (tenantapi.OnboardingTransition, error) {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil ||
		!validReconcileRequest(request) {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingInvalid
	}
	transition, err := coordinator.store.Reconcile(ctx, request)
	if err != nil {
		return tenantapi.OnboardingTransition{}, fmt.Errorf("reconcile onboarding operation: %w", err)
	}
	if !transition.IsValid() || transition.OperationID != request.OperationID {
		return tenantapi.OnboardingTransition{}, tenantapi.ErrOnboardingConflict
	}
	return transition, nil
}

func validReconcileRequest(request ReconcileRequest) bool {
	return request.OperationID != uuid.Nil && request.ReceiptID != uuid.Nil &&
		validDigest(request.ReceiptSHA256) && validDigest(request.EvidenceSHA256) &&
		request.ExpectedRevision > 0
}

package onboarding

import (
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func scanOnboardingTransition(row Row) (tenantapi.OnboardingTransition, error) {
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	if err := row.Scan(&operationID, &state, &revision, &updatedAt); err != nil {
		return tenantapi.OnboardingTransition{}, err
	}
	transition := tenantapi.OnboardingTransition{OperationID: operationID,
		State: tenantapi.OnboardingState(state), Revision: revision,
		UpdatedAt: updatedAt.UTC()}
	if !transition.IsValid() {
		return tenantapi.OnboardingTransition{}, errUnavailable
	}
	return transition, nil
}

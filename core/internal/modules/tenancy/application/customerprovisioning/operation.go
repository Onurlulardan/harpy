package customerprovisioning

import (
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type PublicState string

const (
	StateAccepted             PublicState = "accepted"
	StateMigrating            PublicState = "migrating"
	StateAwaitingAuthorityAck PublicState = "awaiting_authority_ack"
	StateFinalizing           PublicState = "finalizing"
	StateSucceeded            PublicState = "succeeded"
	StateFailed               PublicState = "failed"
)

type Operation struct {
	OperationID, TenantID, RequestedBySubjectID uuid.UUID
	State                                       tenantapi.OnboardingState
	SetupDeliveryState                          string
	FailureCode                                 string
	FailureRetryable                            bool
	CorrelationID                               string
	CreatedAt, UpdatedAt                        time.Time
}

func (operation Operation) PublicState() (PublicState, error) {
	switch operation.State {
	case tenantapi.OnboardingAccepted:
		return StateAccepted, nil
	case tenantapi.OnboardingMigrating:
		return StateMigrating, nil
	case tenantapi.OnboardingMigrationReady, tenantapi.OnboardingAuthorityPending:
		return StateAwaitingAuthorityAck, nil
	case tenantapi.OnboardingAuthorityAcknowledged, tenantapi.OnboardingFinalizing:
		return StateFinalizing, nil
	case tenantapi.OnboardingSucceeded:
		return StateSucceeded, nil
	case tenantapi.OnboardingFailed, tenantapi.OnboardingFenced:
		return StateFailed, nil
	default:
		return "", ErrUnavailable
	}
}

func (operation Operation) valid() bool {
	_, err := operation.PublicState()
	return err == nil && operation.OperationID != uuid.Nil && operation.TenantID != uuid.Nil &&
		operation.RequestedBySubjectID != uuid.Nil && operation.SetupDeliveryState != "" &&
		!operation.CreatedAt.IsZero() && !operation.UpdatedAt.Before(operation.CreatedAt)
}

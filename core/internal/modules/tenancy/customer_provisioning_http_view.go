package tenancy

import (
	"time"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

type customerProvisioningAccepted struct {
	OperationID uuid.UUID `json:"operation_id"`
	Location    string    `json:"location"`
}

type customerProvisioningOperation struct {
	OperationID        uuid.UUID                    `json:"operation_id"`
	State              provisioning.PublicState     `json:"state"`
	TenantID           *uuid.UUID                   `json:"tenant_id,omitempty"`
	AdminSetupDelivery string                       `json:"admin_setup_delivery_state"`
	Failure            *customerProvisioningFailure `json:"failure,omitempty"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}

type customerProvisioningFailure struct {
	Code          string `json:"code"`
	Retryable     bool   `json:"retryable"`
	CorrelationID string `json:"correlation_id"`
}

func customerProvisioningAcceptedView(
	operation provisioning.Operation, location string,
) customerProvisioningAccepted {
	return customerProvisioningAccepted{OperationID: operation.OperationID, Location: location}
}

func customerProvisioningOperationView(
	operation provisioning.Operation,
) customerProvisioningOperation {
	state, _ := operation.PublicState()
	view := customerProvisioningOperation{OperationID: operation.OperationID, State: state,
		AdminSetupDelivery: operation.SetupDeliveryState,
		CreatedAt:          operation.CreatedAt, UpdatedAt: operation.UpdatedAt}
	if state == provisioning.StateSucceeded {
		tenantID := operation.TenantID
		view.TenantID = &tenantID
	}
	if state == provisioning.StateFailed {
		view.Failure = &customerProvisioningFailure{Code: operation.FailureCode,
			Retryable: operation.FailureRetryable, CorrelationID: operation.CorrelationID}
	}
	return view
}

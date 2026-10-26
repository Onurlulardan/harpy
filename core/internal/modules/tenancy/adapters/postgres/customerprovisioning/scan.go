package customerprovisioning

import (
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

func scanOperation(row Row, notFound RowNotFound) (provisioning.Operation, error) {
	var operation provisioning.Operation
	var state string
	var failureCode, correlationID *string
	var failureRetryable *bool
	var createdAt, updatedAt time.Time
	err := row.Scan(&operation.OperationID, &operation.TenantID,
		&operation.RequestedBySubjectID, &state, &operation.SetupDeliveryState,
		&failureCode, &failureRetryable, &correlationID, &createdAt, &updatedAt)
	if err != nil {
		return provisioning.Operation{}, classify(err, notFound)
	}
	operation.State = tenantapi.OnboardingState(state)
	operation.CreatedAt, operation.UpdatedAt = createdAt.UTC(), updatedAt.UTC()
	if failureCode != nil {
		operation.FailureCode = *failureCode
	}
	if failureRetryable != nil {
		operation.FailureRetryable = *failureRetryable
	}
	if correlationID != nil {
		if _, err := uuid.Parse(*correlationID); err != nil {
			return provisioning.Operation{}, provisioning.ErrUnavailable
		}
		operation.CorrelationID = *correlationID
	}
	if _, err := operation.PublicState(); err != nil || operation.OperationID == uuid.Nil ||
		operation.TenantID == uuid.Nil || operation.RequestedBySubjectID == uuid.Nil {
		return provisioning.Operation{}, provisioning.ErrUnavailable
	}
	return operation, nil
}

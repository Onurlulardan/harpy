package customerprovisioning

import (
	"context"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
	"github.com/google/uuid"
)

func (store *Store) LoadSetupReissueTarget(
	ctx context.Context,
	tenantID uuid.UUID,
) (provisioning.SetupReissueTarget, error) {
	if !store.usable(ctx) || tenantID == uuid.Nil {
		return provisioning.SetupReissueTarget{}, provisioning.ErrInvalid
	}
	return store.scanSetupReissueTarget(
		store.queryRow(ctx, getSetupReissueTargetSQL, tenantID),
	)
}

func (store *Store) scanSetupReissueTarget(row Row) (provisioning.SetupReissueTarget, error) {
	var target provisioning.SetupReissueTarget
	var state string
	var failureCode, correlationID *string
	var failureRetryable *bool
	var reissueOperationID *uuid.UUID
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&target.Operation.OperationID, &target.Operation.TenantID,
		&target.Operation.RequestedBySubjectID, &state,
		&target.Operation.SetupDeliveryState, &failureCode,
		&failureRetryable, &correlationID, &createdAt, &updatedAt,
		&reissueOperationID,
		&target.SchemaName, &target.UserID, &target.Email,
	)
	if err != nil {
		return provisioning.SetupReissueTarget{}, classify(err, store.notFound)
	}
	target.Operation.State = tenantapi.OnboardingState(state)
	target.Operation.CreatedAt = createdAt.UTC()
	target.Operation.UpdatedAt = updatedAt.UTC()
	if failureCode != nil {
		target.Operation.FailureCode = *failureCode
	}
	if failureRetryable != nil {
		target.Operation.FailureRetryable = *failureRetryable
	}
	if correlationID != nil {
		if _, err := uuid.Parse(*correlationID); err != nil {
			return provisioning.SetupReissueTarget{}, provisioning.ErrUnavailable
		}
		target.Operation.CorrelationID = *correlationID
	}
	if reissueOperationID != nil {
		target.ReissueOperationID = *reissueOperationID
	}
	if target.Operation.OperationID == uuid.Nil || target.Operation.TenantID == uuid.Nil ||
		target.Operation.RequestedBySubjectID == uuid.Nil || target.UserID == uuid.Nil ||
		target.SchemaName == "" || target.Email == "" {
		return provisioning.SetupReissueTarget{}, provisioning.ErrUnavailable
	}
	return target, nil
}

package onboarding

import (
	"context"
	"fmt"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

func (storage *postgresStore) claimActivation(
	ctx context.Context,
	request onboardingapp.ActivationRequest,
) (tenantapi.ClaimedActivation, error) {
	if !storage.usable(ctx) {
		return tenantapi.ClaimedActivation{}, errUnavailable
	}
	var result tenantapi.ClaimedActivation
	var operationID, tenantID, migrationReceiptID uuid.UUID
	var kind string
	var ack *string
	var databaseNow, leaseUntil time.Time
	err := storage.queryRow(ctx, activateSQL, request.OperationID,
		request.LeaseTokenSHA256, request.ExpectedRevision).Scan(
		&operationID, &tenantID, &kind, &result.StepPlanSHA256,
		&migrationReceiptID, &result.MigrationReceiptSHA256, &ack,
		&result.Revision, &databaseNow, &leaseUntil,
	)
	if err != nil {
		return tenantapi.ClaimedActivation{}, fmt.Errorf("claim activation row: %w", err)
	}
	result.OperationID = operationID
	result.TenantID = tenantID
	result.MigrationReceiptID = migrationReceiptID
	result.Kind = tenantapi.OnboardingKind(kind)
	result.DatabaseNow = databaseNow.UTC()
	result.LeaseUntil = leaseUntil.UTC()
	if ack != nil {
		result.AuthorityAckSHA256 = *ack
	}
	if !result.IsValid() || result.OperationID == uuid.Nil {
		return tenantapi.ClaimedActivation{}, errUnavailable
	}
	return result, nil
}

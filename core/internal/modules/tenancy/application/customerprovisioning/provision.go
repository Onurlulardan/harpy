package customerprovisioning

import (
	"context"
	"errors"
	"strings"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) provision(
	ctx context.Context,
	command Command,
) (Operation, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil ||
		!command.Caller.valid() || !validIdempotencyKey(command.IdempotencyKey) {
		return Operation{}, ErrInvalid
	}
	policy, target, err := coordinator.saaSBoundary()
	if err != nil {
		return Operation{}, err
	}
	requestSHA, err := command.Customer.digest()
	if err != nil {
		return Operation{}, err
	}
	environmentID, environmentErr := uuid.Parse(policy.EnvironmentID())
	controlID, controlErr := uuid.Parse(target.TenantID())
	if environmentErr != nil || controlErr != nil {
		return Operation{}, ErrUnavailable
	}
	now := coordinator.now()
	accessRequest := AccessRequest{Caller: command.Caller, Capability: CapabilityCreate,
		ControlTenantID: controlID, RequestSHA256: requestSHA,
		IdempotencyKey: command.IdempotencyKey, Credential: command.Credential}
	decision, err := coordinator.dependencies.Access.Verify(ctx, accessRequest)
	if err != nil {
		return Operation{}, classifyAccess(err)
	}
	if !decision.validFor(accessRequest, now) || decision.DeploymentMode != policy.Mode() {
		return Operation{}, ErrForbidden
	}
	offering, err := coordinator.dependencies.Offering.ResolveSaaSFreeBase(ctx)
	if err != nil || !validSaaSOffering(offering) {
		return Operation{}, ErrUnavailable
	}
	plan, err := coordinator.buildPlan(ctx, buildPlanInput{
		EnvironmentID: environmentID, IdempotencyKey: command.IdempotencyKey,
		RequestSHA256: requestSHA, Offering: offering, CreatedAt: now,
	})
	if err != nil {
		return Operation{}, err
	}
	err = coordinator.dependencies.Setup.PrepareInitialAdmin(ctx, SetupRequest{
		OperationID: uuid.UUID(plan.OperationID()), TenantID: uuid.UUID(plan.TenantID()),
		UserID: uuid.UUID(plan.FirstUserID()), Email: command.Customer.FirstAdmin.Email,
		CreatedAt: now,
	})
	if err != nil {
		return Operation{}, ErrUnavailable
	}
	operation, err := coordinator.dependencies.Store.Create(ctx, plan, AcceptedRequest{
		RequestedBySubjectID: decision.SubjectID, ControlTenantID: controlID,
		JobID:    stableID(uuid.UUID(plan.OperationID()), "tenant-migration-job"),
		Customer: command.Customer,
	})
	if err != nil {
		return Operation{}, classifyStore(err)
	}
	if !operation.valid() || operation.OperationID != uuid.UUID(plan.OperationID()) ||
		operation.TenantID != uuid.UUID(plan.TenantID()) ||
		operation.RequestedBySubjectID != decision.SubjectID {
		return Operation{}, ErrUnavailable
	}
	return operation, nil
}

func (coordinator *Coordinator) now() time.Time {
	return coordinator.dependencies.Clock.Now().UTC().Truncate(time.Microsecond)
}

func validIdempotencyKey(value string) bool {
	return len(value) >= 1 && len(value) <= 128 && strings.TrimSpace(value) == value
}

func validSaaSOffering(value tenantapi.BaseOfferingPin) bool {
	return value.VersionID != tenantapi.NilUUID && value.Revision == 1 &&
		len(value.SHA256) == 64 && !value.EffectiveAt.IsZero() &&
		value.EffectiveAt.Location() == time.UTC && !value.Source.Present
}

func classifyAccess(err error) error {
	switch {
	case errors.Is(err, ErrAuthentication):
		return ErrAuthentication
	case errors.Is(err, ErrForbidden):
		return ErrForbidden
	default:
		return ErrUnavailable
	}
}

func classifyStore(err error) error {
	switch {
	case errors.Is(err, ErrConflict):
		return ErrConflict
	case errors.Is(err, ErrNotFound):
		return ErrNotFound
	default:
		return ErrUnavailable
	}
}

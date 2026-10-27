package customerprovisioning

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type ReadQuery struct {
	Caller      CallerKind
	OperationID uuid.UUID
	Credential  CredentialEnvelope
}

func (coordinator *Coordinator) Read(
	ctx context.Context,
	query ReadQuery,
) (Operation, error) {
	if coordinator == nil || ctx == nil || ctx.Err() != nil ||
		!query.Caller.valid() || query.OperationID == uuid.Nil {
		return Operation{}, ErrInvalid
	}
	policy, target, err := coordinator.saaSBoundary()
	if err != nil {
		return Operation{}, err
	}
	controlID, _ := uuid.Parse(target.TenantID())
	accessRequest := AccessRequest{Caller: query.Caller, Capability: CapabilityRead,
		ControlTenantID: controlID, Credential: query.Credential}
	decision, err := coordinator.dependencies.Access.Verify(ctx, accessRequest)
	if err != nil {
		return Operation{}, classifyAccess(err)
	}
	if !decision.validFor(accessRequest, coordinator.now()) ||
		decision.DeploymentMode != policy.Mode() {
		return Operation{}, ErrForbidden
	}
	operation, err := coordinator.dependencies.Store.Read(
		ctx, query.OperationID, decision.SubjectID,
	)
	if err != nil {
		return Operation{}, err
	}
	if !operation.valid() || operation.RequestedBySubjectID != decision.SubjectID {
		return Operation{}, ErrNotFound
	}
	return operation, nil
}

func (coordinator *Coordinator) saaSBoundary() (
	tenantapi.DeploymentPolicy, tenantapi.PrimaryTenantTarget, error,
) {
	policy, err := coordinator.dependencies.Policy.Snapshot()
	if err != nil || !policy.IsValid() {
		return tenantapi.DeploymentPolicy{}, tenantapi.PrimaryTenantTarget{}, ErrUnavailable
	}
	if policy.Mode() != tenantapi.DeploymentModeSaaS {
		return tenantapi.DeploymentPolicy{}, tenantapi.PrimaryTenantTarget{}, ErrModeDenied
	}
	target, err := coordinator.dependencies.Policy.PrimaryTarget()
	if err != nil || !target.IsValid() ||
		target.DeploymentMode() != tenantapi.DeploymentModeSaaS ||
		target.Purpose() != tenantapi.PrimaryTenantPurposeSaaS ||
		target.EnvironmentID() != policy.EnvironmentID() {
		return tenantapi.DeploymentPolicy{}, tenantapi.PrimaryTenantTarget{}, ErrUnavailable
	}
	return policy, target, nil
}

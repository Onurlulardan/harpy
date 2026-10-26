package environmentoperator

import (
	"context"
	"errors"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func (coordinator *Coordinator) prepareAuthority(
	ctx context.Context,
	request AuthorityRequest,
) (tenantapi.ProtectedAuthorityPlan, error) {
	if request.DeploymentMode == tenantapi.DeploymentModeOnPremise {
		return tenantapi.ProtectedAuthorityPlan{
			RegistryBindingID: request.RegistryBindingID,
			ScopeKind:         tenantapi.AuthorityScopeEnvironment,
			ScopeID:           request.EnvironmentID,
		}, nil
	}
	if coordinator.authority == nil {
		return tenantapi.ProtectedAuthorityPlan{}, ErrDependencyMismatch
	}
	plan, err := coordinator.authority.Prepare(ctx, request)
	if err != nil {
		return tenantapi.ProtectedAuthorityPlan{}, classifyDependency(err)
	}
	if plan.RegistryBindingID != request.RegistryBindingID ||
		plan.ScopeKind != tenantapi.AuthorityScopeEnvironment ||
		plan.ScopeID != request.EnvironmentID || plan.Generation != 1 ||
		plan.PreviousHeadSHA256 != "" {
		return tenantapi.ProtectedAuthorityPlan{}, ErrFenced
	}
	return plan, nil
}

func (coordinator *Coordinator) acknowledgeAuthority(
	ctx context.Context,
	prepared preparedPlan,
	revision int64,
) (int64, error) {
	if coordinator.authority == nil {
		return 0, ErrDependencyMismatch
	}
	ack, err := coordinator.authority.Acknowledge(ctx, prepared.authorityRequest)
	if err != nil {
		return 0, classifyDependency(err)
	}
	plan := prepared.plan.Authority()
	if ack.OperationID != prepared.plan.OperationID() ||
		ack.RegistryBindingID != plan.RegistryBindingID ||
		ack.ScopeKind != plan.ScopeKind || ack.ScopeID != plan.ScopeID ||
		ack.Generation != plan.Generation ||
		ack.PreviousHeadSHA256 != plan.PreviousHeadSHA256 ||
		ack.StateSHA256 != plan.StateSHA256 ||
		ack.ProposalSHA256 != plan.ProposalSHA256 {
		return 0, ErrFenced
	}
	transition, err := coordinator.onboarding.ReceiveAuthorityAck(ctx, ack, revision)
	if err != nil {
		if errors.Is(err, tenantapi.ErrAuthorityCASConflict) {
			return 0, ErrFenced
		}
		return 0, ErrResumable
	}
	return transition.Revision, nil
}

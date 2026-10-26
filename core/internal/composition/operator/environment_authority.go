package operator

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
)

type EnvironmentAuthority struct {
	planner      platformauthority.InitialAuthorityPlanner
	acknowledger platformauthority.InitialAuthorityAcknowledger
	binding      platformauthority.SaaSBinding
}

func NewEnvironmentAuthority(
	planner platformauthority.InitialAuthorityPlanner,
	acknowledger platformauthority.InitialAuthorityAcknowledger,
	binding platformauthority.SaaSBinding,
) (*EnvironmentAuthority, error) {
	if planner == nil || acknowledger == nil || !binding.Valid() {
		return nil, fmt.Errorf("SaaS environment authority is unavailable")
	}
	return &EnvironmentAuthority{planner: planner,
		acknowledger: acknowledger, binding: binding}, nil
}

func (authority *EnvironmentAuthority) Prepare(
	ctx context.Context,
	request environmentoperator.AuthorityRequest,
) (tenancymodule.ProtectedAuthorityPlan, error) {
	if authority == nil || request.DeploymentMode != tenancymodule.DeploymentModeSaaS ||
		request.RegistryBindingID.String() != authority.binding.BindingID ||
		request.EnvironmentID.String() != authority.binding.EnvironmentID {
		return tenancymodule.ProtectedAuthorityPlan{}, fmt.Errorf("SaaS authority request is invalid")
	}
	ownerID := request.InstanceOwnerBindingID.String()
	ownerGeneration := int64(1)
	adminSHA := iammodule.ProtectedAdminBindingSetSHA256(
		request.OperationID, request.TenantID, request.FirstUserID,
		request.ProtectedAdminBindingID,
	)
	state := platformauthority.AuthorityState{Schema: platformauthority.AuthorityStateSchema,
		EnvironmentID:            request.EnvironmentID.String(),
		DeploymentMode:           platformauthority.AuthorityModeSaaS,
		RegistryBindingID:        authority.binding.BindingID,
		ScopeKind:                platformauthority.AuthorityScopeEnvironment,
		ScopeID:                  request.EnvironmentID.String(),
		ProposalSignerGeneration: authority.binding.SignerGeneration,
		ProposalSignerKID:        authority.binding.SignerKID,
		OwnerBindingID:           &ownerID, OwnerBindingGeneration: &ownerGeneration,
		PrimaryAdminBindingSetSHA256: &adminSHA}
	prepared, err := authority.planner.PrepareInitial(ctx,
		platformauthority.InitialAuthorityRequest{OperationID: request.OperationID.String(),
			State: state, IssuedAt: request.IssuedAt, Policy: platformauthority.AuthorityExpiryOnline})
	if err != nil {
		return tenancymodule.ProtectedAuthorityPlan{}, err
	}
	if prepared.OperationID != request.OperationID.String() ||
		prepared.RegistryBindingID != authority.binding.BindingID ||
		prepared.ScopeKind != string(platformauthority.AuthorityScopeEnvironment) ||
		prepared.ScopeID != request.EnvironmentID.String() ||
		prepared.Generation != 1 || prepared.StateSHA256 == "" ||
		prepared.ProposalSHA256 == "" {
		return tenancymodule.ProtectedAuthorityPlan{}, fmt.Errorf("SaaS authority proposal differs")
	}
	return tenancymodule.ProtectedAuthorityPlan{RegistryBindingID: request.RegistryBindingID,
		ScopeKind: tenancymodule.AuthorityScopeEnvironment, ScopeID: request.EnvironmentID,
		Generation: 1, StateSHA256: prepared.StateSHA256,
		ProposalSHA256: prepared.ProposalSHA256}, nil
}

func (authority *EnvironmentAuthority) Acknowledge(
	ctx context.Context,
	request environmentoperator.AuthorityRequest,
) (tenancymodule.AuthorityAcknowledgement, error) {
	if authority == nil {
		return tenancymodule.AuthorityAcknowledgement{}, fmt.Errorf("SaaS authority is unavailable")
	}
	receipt, err := authority.acknowledger.AcknowledgePrepared(ctx, request.OperationID.String())
	if err != nil {
		return tenancymodule.AuthorityAcknowledgement{}, err
	}
	bindingID, bindingErr := tenancymodule.ParseUUID(receipt.RegistryBindingID)
	scopeID, scopeErr := tenancymodule.ParseUUID(receipt.ScopeID)
	if bindingErr != nil || scopeErr != nil || receipt.OperationID != request.OperationID.String() {
		return tenancymodule.AuthorityAcknowledgement{}, fmt.Errorf("SaaS authority receipt differs")
	}
	return tenancymodule.AuthorityAcknowledgement{OperationID: request.OperationID,
		RegistryBindingID: bindingID, ScopeKind: tenancymodule.AuthorityScopeEnvironment,
		ScopeID: scopeID, Generation: receipt.Generation,
		StateSHA256: receipt.StateSHA256, ProposalSHA256: receipt.ProposalSHA256,
		AckSHA256: receipt.AckSHA256, RegistrySequence: receipt.RegistrySequence,
		VerifiedAt: receipt.CompletedAt}, nil
}

var _ environmentoperator.ProtectedAuthority = (*EnvironmentAuthority)(nil)

package api

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type customerProvisioningAuthority struct {
	planner platformauthority.InitialAuthorityPlanner
	binding platformauthority.SaaSBinding
}

func (authority customerProvisioningAuthority) PrepareInitialTenant(
	ctx context.Context,
	request tenancymodule.CustomerProvisioningAuthorityRequest,
) (tenancymodule.ProtectedAuthorityPlan, error) {
	if authority.planner == nil || !authority.binding.Valid() {
		return tenancymodule.ProtectedAuthorityPlan{}, fmt.Errorf("SaaS authority planner unavailable")
	}
	tenantID := request.TenantID.String()
	adminSHA := iammodule.ProtectedAdminBindingSetSHA256(
		request.OperationID, request.TenantID, request.FirstUserID,
		request.ProtectedAdminBindingID,
	)
	state := platformauthority.AuthorityState{Schema: platformauthority.AuthorityStateSchema,
		EnvironmentID: request.EnvironmentID.String(), DeploymentMode: platformauthority.AuthorityModeSaaS,
		RegistryBindingID: authority.binding.BindingID,
		ScopeKind:         platformauthority.AuthorityScopeTenant, ScopeID: tenantID,
		ProposalSignerGeneration: authority.binding.SignerGeneration,
		ProposalSignerKID:        authority.binding.SignerKID,
		TenantID:                 &tenantID, ProtectedAdminBindingSetSHA256: &adminSHA}
	prepared, err := authority.planner.PrepareInitial(ctx, platformauthority.InitialAuthorityRequest{
		OperationID: request.OperationID.String(), State: state,
		IssuedAt: request.IssuedAt, Policy: platformauthority.AuthorityExpiryOnline,
	})
	if err != nil {
		return tenancymodule.ProtectedAuthorityPlan{}, err
	}
	bindingID, bindingErr := tenancymodule.ParseUUID(prepared.RegistryBindingID)
	scopeID, scopeErr := tenancymodule.ParseUUID(prepared.ScopeID)
	if bindingErr != nil || scopeErr != nil || prepared.ScopeKind != string(platformauthority.AuthorityScopeTenant) ||
		prepared.Generation != 1 || prepared.StateSHA256 == "" || prepared.ProposalSHA256 == "" {
		return tenancymodule.ProtectedAuthorityPlan{}, fmt.Errorf("SaaS authority proposal is invalid")
	}
	return tenancymodule.ProtectedAuthorityPlan{RegistryBindingID: bindingID,
		ScopeKind: tenancymodule.AuthorityScopeTenant, ScopeID: scopeID, Generation: 1,
		StateSHA256: prepared.StateSHA256, ProposalSHA256: prepared.ProposalSHA256}, nil
}

var _ tenancymodule.CustomerProvisioningAuthorityPlanner = customerProvisioningAuthority{}

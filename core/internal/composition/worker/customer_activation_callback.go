package worker

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func newCustomerActivationCallback(
	directory config.IAMDirectoryKeyring,
) (tenancy.CustomerActivationCallback, error) {
	if directory.Validate() != nil {
		return nil, fmt.Errorf("customer activation directory keyring is invalid")
	}
	return func(ctx context.Context, claim tenancy.CustomerWorkerClaim,
		binding tenancy.OperatorActivationBinding,
		resources tenancy.OperatorActivationResources,
	) error {
		if !validCustomerActivation(claim, binding, resources) {
			return fmt.Errorf("customer activation binding is invalid")
		}
		artifactID, err := finalizeCustomerTenantIdentity(
			ctx, claim, binding, resources, directory,
		)
		if err != nil {
			return fmt.Errorf("finalize customer identity: %w", err)
		}
		if err := finalizeCustomerBaseline(ctx, claim, binding, resources); err != nil {
			return fmt.Errorf("finalize customer baseline: %w", err)
		}
		if err := finalizeCustomerEvidence(
			ctx, binding, resources, artifactID,
		); err != nil {
			return fmt.Errorf("finalize customer evidence: %w", err)
		}
		return nil
	}, nil
}

func validCustomerActivation(claim tenancy.CustomerWorkerClaim,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) bool {
	plan := claim.Plan
	authority := plan.Authority()
	return claim.Valid() && resources.BindOwner != nil && resources.Consume != nil &&
		binding.Kind == "saas_customer_tenant_provision" &&
		binding.DeploymentMode == "saas" && binding.BindingPurpose == "saas_customer" &&
		binding.InstanceOwnerBindingID == [16]byte{} &&
		binding.OperationID == plan.OperationID() && binding.TenantID == plan.TenantID() &&
		binding.FirstUserID == plan.FirstUserID() &&
		binding.AuthorityScopeKind == "tenant" &&
		binding.AuthorityScopeID == binding.TenantID &&
		binding.AuthorityRegistryBindingID == authority.RegistryBindingID &&
		binding.AuthorityGeneration == 1 && binding.AuthorityPreviousHeadSHA256 == "" &&
		binding.AuthorityStateSHA256 == authority.StateSHA256 &&
		binding.AuthorityProposalSHA256 == authority.ProposalSHA256 &&
		validCustomerActivationSHA256(binding.AuthorityAckSHA256)
}

func validCustomerActivationSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && value == strings.ToLower(value)
}

func consumeCustomer(resources tenancy.OperatorActivationResources,
	command tenancy.OperatorActivationCommand,
) error {
	if err := resources.Consume(command); err != nil {
		return fmt.Errorf("consume %s: %w", command, err)
	}
	return nil
}

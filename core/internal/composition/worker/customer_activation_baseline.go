package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func finalizeCustomerBaseline(ctx context.Context, _ tenancy.CustomerWorkerClaim,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) error {
	owner, err := resources.BindOwner(tenancy.OperatorOwnerBilling)
	if err != nil {
		return fmt.Errorf("bind billing owner: %w", err)
	}
	executor := customerBillingExecutor(owner)
	if err = billing.AssignInstallationBaseOffering(ctx, executor,
		billing.InstallationBaseOfferingInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, OfferingVersionID: binding.OfferingVersionID,
			DeploymentMode: binding.DeploymentMode, BindingPurpose: binding.BindingPurpose,
			SchemaName: binding.SchemaName, OfferingRevision: binding.OfferingRevision,
			OfferingSHA256:      binding.OfferingSHA256,
			OfferingEffectiveAt: binding.OfferingEffectiveAt,
			ManifestSHA256:      binding.ManifestSHA256,
		}); err != nil {
		return err
	}
	if err = consumeCustomer(resources, tenancy.OperatorAssignTenantBaseOffering); err != nil {
		return err
	}
	if err = billing.OpenInstallationUsage(ctx, executor,
		billing.InstallationUsageInput{
			OperationID: binding.OperationID, TenantID: binding.TenantID,
			ManifestSHA256: binding.ManifestSHA256, SchemaName: binding.SchemaName,
		}); err != nil {
		return err
	}
	return consumeCustomer(resources, tenancy.OperatorOpenTenantUsage)
}

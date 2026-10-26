package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresourcelayout "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func finalizeCustomerBaseline(ctx context.Context, claim tenancy.CustomerWorkerClaim,
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
	if err = consumeCustomer(resources, tenancy.OperatorOpenTenantUsage); err != nil {
		return err
	}
	return initializeCustomerSystemLayout(ctx, claim, binding, resources)
}

func initializeCustomerSystemLayout(ctx context.Context, claim tenancy.CustomerWorkerClaim,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) error {
	catalogs, err := iam.BuildInstallationPresentationCatalogs(binding.OperationID)
	if err != nil {
		return err
	}
	user, err := catalogs.User.Presentation()
	if err != nil {
		return err
	}
	organization, err := catalogs.Organization.Presentation()
	if err != nil {
		return err
	}
	role, err := catalogs.Role.Presentation()
	if err != nil {
		return err
	}
	owner, err := resources.BindOwner(tenancy.OperatorOwnerSystemResources)
	if err != nil {
		return fmt.Errorf("bind system resource owner: %w", err)
	}
	err = systemresourcelayout.InitializeInstallationBaseline(ctx,
		customerLayoutExecutor(owner), systemresourcelayout.InstallationBaselineInput{
			OperationID: binding.OperationID, TenantID: binding.TenantID,
			ActivatedAt: claim.Plan.CreatedAt(), SchemaName: binding.SchemaName,
			UserCatalog: user, OrganizationCatalog: organization, RoleCatalog: role,
		})
	if err != nil {
		return err
	}
	return consumeCustomer(resources, tenancy.OperatorInitializeSystemBaseline)
}

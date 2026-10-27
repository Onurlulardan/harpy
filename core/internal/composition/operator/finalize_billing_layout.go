package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresourcelayout "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func finalizeBillingAndLayout(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	input EnvironmentFinalizerInput,
) error {
	billingOwner, err := resources.BindOwner(tenancy.OperatorOwnerBilling)
	if err != nil {
		return fmt.Errorf("bind billing owner: %w", err)
	}
	executor := billingExecutor(billingOwner)
	if err := billing.AssignInstallationBaseOffering(ctx, executor,
		billing.InstallationBaseOfferingInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, OfferingVersionID: binding.OfferingVersionID,
			DeploymentMode: binding.DeploymentMode, BindingPurpose: binding.BindingPurpose,
			SchemaName: binding.SchemaName, OfferingRevision: binding.OfferingRevision,
			OfferingSHA256:      binding.OfferingSHA256,
			OfferingEffectiveAt: binding.OfferingEffectiveAt,
			Source: billing.InstallationOfferingSource{
				Sequence: binding.OfferingSourceSequence,
				SHA256:   binding.OfferingSourceSHA256,
				Present:  binding.OfferingSourcePresent,
			}, ManifestSHA256: binding.ManifestSHA256,
		},
	); err != nil {
		return err
	}
	if err := consume(resources, tenancy.OperatorAssignTenantBaseOffering); err != nil {
		return err
	}
	if err := billing.OpenInstallationUsage(ctx, executor,
		billing.InstallationUsageInput{
			OperationID: binding.OperationID, TenantID: binding.TenantID,
			ManifestSHA256: binding.ManifestSHA256, SchemaName: binding.SchemaName,
		},
	); err != nil {
		return err
	}
	if err := consume(resources, tenancy.OperatorOpenTenantUsage); err != nil {
		return err
	}
	return finalizeSystemLayout(ctx, binding, resources, input.ActivatedAt)
}

func finalizeSystemLayout(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	activatedAt time.Time,
) error {
	catalogs, err := iam.BuildInstallationPresentationCatalogs(binding.OperationID)
	if err != nil {
		return err
	}
	userCatalog, err := catalogs.User.Presentation()
	if err != nil {
		return err
	}
	organizationCatalog, err := catalogs.Organization.Presentation()
	if err != nil {
		return err
	}
	roleCatalog, err := catalogs.Role.Presentation()
	if err != nil {
		return err
	}
	owner, err := resources.BindOwner(tenancy.OperatorOwnerSystemResources)
	if err != nil {
		return fmt.Errorf("bind system resource owner: %w", err)
	}
	err = systemresourcelayout.InitializeInstallationBaseline(ctx,
		layoutExecutor(owner),
		systemresourcelayout.InstallationBaselineInput{
			OperationID: binding.OperationID, TenantID: binding.TenantID,
			ActivatedAt: activatedAt, SchemaName: binding.SchemaName,
			UserCatalog: userCatalog, OrganizationCatalog: organizationCatalog,
			RoleCatalog: roleCatalog,
		},
	)
	if err != nil {
		return err
	}
	return consume(resources, tenancy.OperatorInitializeSystemBaseline)
}

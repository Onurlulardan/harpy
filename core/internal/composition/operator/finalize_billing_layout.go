package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func finalizeBillingAndLayout(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	_ EnvironmentFinalizerInput,
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
	return consume(resources, tenancy.OperatorOpenTenantUsage)
}

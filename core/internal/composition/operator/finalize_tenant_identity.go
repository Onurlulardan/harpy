package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

func finalizeTenantAndIdentity(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	input EnvironmentFinalizerInput,
) error {
	tenantOwner, err := resources.BindOwner(tenancy.OperatorOwnerTenancy)
	if err != nil {
		return fmt.Errorf("bind tenancy owner: %w", err)
	}
	if err := tenancy.ActivateTenant(ctx, tenantOwner, tenancy.ActivateTenantInput{
		OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
		TenantID: binding.TenantID, OwnerUserID: binding.FirstUserID,
		Name: input.TenantName, Slug: input.TenantSlug,
		SchemaName: binding.SchemaName, Kind: binding.Kind,
		BindingPurpose: binding.BindingPurpose,
		ManifestSHA256: binding.ManifestSHA256,
	}); err != nil {
		return err
	}
	if err := consume(resources, tenancy.OperatorActivateTenant); err != nil {
		return err
	}
	iamOwner, err := resources.BindOwner(tenancy.OperatorOwnerIAM)
	if err != nil {
		return fmt.Errorf("bind IAM owner: %w", err)
	}
	if err := iam.InitializeInstallationIdentity(ctx,
		iamExecutor(iamOwner),
		iam.InstallationIdentityInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, FirstUserID: binding.FirstUserID,
			RootOrganizationID:      binding.RootOrganizationID,
			ProtectedAdminBindingID: binding.ProtectedAdminBindingID,
			Email:                   input.AdminEmail, DisplayName: input.AdminDisplayName,
			TenantName: input.TenantName, SchemaName: binding.SchemaName,
			DirectoryKeyID: input.DirectoryKeyID,
			IdentifierHMAC: input.IdentifierHMAC,
			SourceSHA256:   input.DirectorySourceSHA256,
		},
	); err != nil {
		return err
	}
	if err := iam.InitializeInstallationCredential(ctx,
		iamExecutor(iamOwner), iam.InstallationCredentialInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, UserID: binding.FirstUserID,
			Credential: input.InitialCredential,
		},
	); err != nil {
		return err
	}
	return consume(resources, tenancy.OperatorInitializeTenantIdentity)
}

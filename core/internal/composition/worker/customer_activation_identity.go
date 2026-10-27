package worker

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func finalizeCustomerTenantIdentity(ctx context.Context, claim tenancy.CustomerWorkerClaim,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	directory config.IAMDirectoryKeyring,
) (iam.UUID, error) {
	tenantOwner, err := resources.BindOwner(tenancy.OperatorOwnerTenancy)
	if err != nil {
		return iam.UUID{}, fmt.Errorf("bind tenancy owner: %w", err)
	}
	if err = tenancy.ActivateTenant(ctx, tenantOwner, tenancy.ActivateTenantInput{
		OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
		TenantID: binding.TenantID, OwnerUserID: binding.FirstUserID,
		Name: claim.Customer.Name, Slug: claim.Customer.Slug,
		SchemaName: binding.SchemaName, Kind: binding.Kind,
		BindingPurpose: binding.BindingPurpose,
		ManifestSHA256: binding.ManifestSHA256,
	}); err != nil {
		return iam.UUID{}, err
	}
	if err = consumeCustomer(resources, tenancy.OperatorActivateTenant); err != nil {
		return iam.UUID{}, err
	}
	keyID, digest, source, err := customerDirectoryEvidence(
		claim.Customer.FirstAdmin.Email, directory,
	)
	if err != nil {
		return iam.UUID{}, err
	}
	iamOwner, err := resources.BindOwner(tenancy.OperatorOwnerIAM)
	if err != nil {
		return iam.UUID{}, fmt.Errorf("bind IAM owner: %w", err)
	}
	executor := customerIAMExecutor(iamOwner)
	if err = iam.InitializeInstallationIdentity(ctx, executor,
		iam.InstallationIdentityInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, FirstUserID: binding.FirstUserID,
			RootOrganizationID:      binding.RootOrganizationID,
			ProtectedAdminBindingID: binding.ProtectedAdminBindingID,
			Email:                   claim.Customer.FirstAdmin.Email,
			DisplayName:             claim.Customer.FirstAdmin.DisplayName,
			TenantName:              claim.Customer.Name, SchemaName: binding.SchemaName,
			DirectoryKeyID: keyID, IdentifierHMAC: digest, SourceSHA256: source,
		}); err != nil {
		return iam.UUID{}, err
	}
	artifactID, err := iam.FinalizeInstallationInitialAdminSetup(ctx, executor,
		iam.InitialAdminSetupFinalizeInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, SchemaName: binding.SchemaName,
			ManifestSHA256: binding.ManifestSHA256,
		})
	if err != nil {
		return iam.UUID{}, err
	}
	return artifactID, consumeCustomer(resources, tenancy.OperatorInitializeTenantIdentity)
}

func customerDirectoryEvidence(email string,
	directory config.IAMDirectoryKeyring,
) (string, [32]byte, [32]byte, error) {
	value, err := iam.NormalizeEmailV1(email)
	keyID, key, keyErr := directory.Active()
	if err != nil || keyErr != nil || value != email {
		clear(key[:])
		return "", [32]byte{}, [32]byte{}, fmt.Errorf("customer email is not canonical")
	}
	mac := hmac.New(sha256.New, key[:])
	_, _ = mac.Write([]byte(value))
	clear(key[:])
	var digest [32]byte
	copy(digest[:], mac.Sum(nil))
	return keyID, digest, sha256.Sum256([]byte(value)), nil
}

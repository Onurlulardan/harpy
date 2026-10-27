package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func finalizePlatformEvidence(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) error {
	if err := finalizeRecoveryIdentity(ctx, binding, resources); err != nil {
		return err
	}
	if err := finalizeAuthority(ctx, binding, resources); err != nil {
		return err
	}
	owner, err := resources.BindOwner(tenancy.OperatorOwnerAsync)
	if err != nil {
		return fmt.Errorf("bind async owner: %w", err)
	}
	return platformasync.RecordInstallationActivationEvent(ctx,
		asyncExecutor(owner),
		platformasync.InstallationActivationEventInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, ActorID: binding.FirstUserID,
			SchemaName: binding.SchemaName, ManifestSHA256: binding.ManifestSHA256,
		},
	)
}

func finalizeRecoveryIdentity(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) error {
	owner, err := resources.BindOwner(tenancy.OperatorOwnerPlatformOps)
	if err != nil {
		return fmt.Errorf("bind platform operations owner: %w", err)
	}
	err = platformops.InitializeInstallationRecoveryIdentity(ctx,
		platformOpsExecutor(owner),
		platformops.InstallationRecoveryIdentityInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, SchemaName: binding.SchemaName,
			DeploymentMode:         binding.DeploymentMode,
			SigningGeneration:      binding.RecoverySigningGeneration,
			TrustBundleSHA256:      binding.RecoveryTrustBundleSHA256,
			RecoverySigningPresent: binding.RecoverySigningPresent,
			ManifestSHA256:         binding.ManifestSHA256,
		},
	)
	if err != nil {
		return err
	}
	return consume(resources, tenancy.OperatorInitializeRecoverySigning)
}

func finalizeAuthority(
	ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
) error {
	owner, err := resources.BindOwner(tenancy.OperatorOwnerPlatformAuthority)
	if err != nil {
		return fmt.Errorf("bind platform authority owner: %w", err)
	}
	executor := authorityExecutor(owner)
	if err := platformauthority.BindInstallationInstanceOwner(ctx, executor,
		platformauthority.InstallationOwnerBindingInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, SchemaName: binding.SchemaName,
			DeploymentMode:          binding.DeploymentMode,
			InstanceOwnerBindingID:  binding.InstanceOwnerBindingID,
			ProtectedAdminBindingID: binding.ProtectedAdminBindingID,
			RegistryBindingID:       binding.AuthorityRegistryBindingID,
			ManifestSHA256:          binding.ManifestSHA256,
		},
	); err != nil {
		return err
	}
	iamOwner, err := resources.BindOwner(tenancy.OperatorOwnerIAM)
	if err != nil {
		return fmt.Errorf("bind IAM platform owner: %w", err)
	}
	if err := iam.InitializeInstallationPlatformOwner(ctx, iamExecutor(iamOwner),
		iam.InstallationPlatformOwnerInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, FirstUserID: binding.FirstUserID,
			InstanceOwnerBindingID: binding.InstanceOwnerBindingID,
			SchemaName:             binding.SchemaName, RequestSHA256: binding.StepPlanSHA256,
		},
	); err != nil {
		return err
	}
	if err := consume(resources, tenancy.OperatorBindInstanceOwner); err != nil {
		return err
	}
	if binding.DeploymentMode == "onpremise" {
		return consume(resources, tenancy.OperatorPinAuthorityHead)
	}
	if err := platformauthority.PinInstallationAuthorityHead(ctx, executor,
		platformauthority.InstallationHeadPinInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, SchemaName: binding.SchemaName,
			RegistryBindingID: binding.AuthorityRegistryBindingID,
			ScopeKind:         binding.AuthorityScopeKind, ScopeID: binding.AuthorityScopeID,
			Generation:         binding.AuthorityGeneration,
			PreviousHeadSHA256: binding.AuthorityPreviousHeadSHA256,
			StateSHA256:        binding.AuthorityStateSHA256,
			ProposalSHA256:     binding.AuthorityProposalSHA256,
			AckSHA256:          binding.AuthorityAckSHA256,
			ManifestSHA256:     binding.ManifestSHA256,
		},
	); err != nil {
		return err
	}
	return consume(resources, tenancy.OperatorPinAuthorityHead)
}

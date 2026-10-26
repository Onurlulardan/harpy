package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func finalizeCustomerEvidence(ctx context.Context,
	binding tenancy.OperatorActivationBinding,
	resources tenancy.OperatorActivationResources,
	artifactID iam.UUID,
) error {
	owner, err := resources.BindOwner(tenancy.OperatorOwnerPlatformAuthority)
	if err != nil {
		return fmt.Errorf("bind platform authority owner: %w", err)
	}
	if err = platformauthority.PinInstallationAuthorityHead(ctx,
		customerAuthorityExecutor(owner), platformauthority.InstallationHeadPinInput{
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
		}); err != nil {
		return err
	}
	if err = consumeCustomer(resources, tenancy.OperatorPinAuthorityHead); err != nil {
		return err
	}
	owner, err = resources.BindOwner(tenancy.OperatorOwnerAsync)
	if err != nil {
		return fmt.Errorf("bind async owner: %w", err)
	}
	executor := customerAsyncExecutor(owner)
	if err = platformasync.EnqueueInitialAdminSetupDelivery(ctx, executor,
		platformasync.InitialAdminSetupDeliveryInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, ArtifactID: artifactID,
			SchemaName:     binding.SchemaName,
			ManifestSHA256: binding.ManifestSHA256,
		}); err != nil {
		return err
	}
	return platformasync.RecordInstallationActivationEvent(ctx, executor,
		platformasync.InstallationActivationEventInput{
			OperationID: binding.OperationID, EnvironmentID: binding.EnvironmentID,
			TenantID: binding.TenantID, ActorID: binding.FirstUserID,
			SchemaName: binding.SchemaName, ManifestSHA256: binding.ManifestSHA256,
		})
}

package onboarding

import (
	"context"
	"fmt"

	activation "github.com/Onurlulardan/harpy/core/internal/capabilities/tenantactivation/api"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) Activate(
	ctx context.Context,
	request ActivationRequest,
	completion ActivationCompletion,
	callback ActivationCallback,
) error {
	if !coordinator.usable() || ctx == nil || ctx.Err() != nil || callback == nil ||
		!validActivationRequest(request, completion) {
		return tenantapi.ErrOnboardingInvalid
	}
	return coordinator.unit.Run(ctx, func(
		transactionContext context.Context,
		transaction ActivationTransaction,
	) error {
		return transaction.Activate(transactionContext, request, completion, func(
			claim tenantapi.ClaimedActivation,
			resources ActivationResources,
		) error {
			if !claim.IsValid() || claim.OperationID != request.OperationID ||
				claim.TenantID != request.Plan.TenantID() || claim.Kind != request.Plan.Kind() ||
				claim.StepPlanSHA256 != request.Plan.StepPlanSHA256() {
				return tenantapi.ErrOnboardingConflict
			}
			if resources.BindTenantSchema == nil || resources.BindOwner == nil {
				return tenantapi.ErrOnboardingInvalid
			}
			if err := resources.BindTenantSchema(
				transactionContext,
				request.Plan.SchemaName(),
			); err != nil {
				return fmt.Errorf("bind activation tenant schema: %w", err)
			}
			grant := activationGrant(request.Plan, claim)
			return activation.Run(grant, claim.DatabaseNow, func(scope activation.Scope) error {
				if err := callback(transactionContext, scope, resources); err != nil {
					return err
				}
				return nil
			})
		})
	})
}

func activationGrant(
	plan tenantapi.OnboardingPlan,
	claim tenantapi.ClaimedActivation,
) activation.Grant {
	offering := plan.Offering()
	recovery := plan.RecoverySigning()
	authority := plan.Authority()
	binding := activation.Binding{OperationID: plan.OperationID(),
		EnvironmentID: plan.EnvironmentID(), TenantID: plan.TenantID(),
		FirstUserID: plan.FirstUserID(), RootOrganizationID: plan.RootOrganizationID(),
		ProtectedAdminBindingID: plan.ProtectedAdminBindingID(),
		InstanceOwnerBindingID:  plan.InstanceOwnerBindingID(), Kind: string(plan.Kind()),
		DeploymentMode: string(plan.DeploymentMode()), BindingPurpose: string(plan.BindingPurpose()),
		SchemaName: plan.SchemaName(), RequestSHA256: plan.RequestSHA256(),
		ManifestSHA256: plan.ManifestSHA256(), StepPlanSHA256: plan.StepPlanSHA256(),
		Offering: activation.OfferingPin{VersionID: offering.VersionID,
			Revision: offering.Revision, SHA256: offering.SHA256, EffectiveAt: offering.EffectiveAt,
			Source: activation.OfferingSourcePin{Sequence: offering.Source.Sequence,
				SHA256: offering.Source.SHA256, Present: offering.Source.Present}},
		RecoverySigning: activation.RecoverySigningPin{Generation: recovery.Generation,
			TrustBundleSHA256: recovery.TrustBundleSHA256, Present: recovery.Present},
		Authority: activation.AuthorityPin{RegistryBindingID: authority.RegistryBindingID,
			ScopeKind: string(authority.ScopeKind), ScopeID: authority.ScopeID,
			Generation: authority.Generation, PreviousHeadSHA256: authority.PreviousHeadSHA256,
			StateSHA256: authority.StateSHA256, ProposalSHA256: authority.ProposalSHA256,
			AckSHA256: claim.AuthorityAckSHA256},
		MigrationReceipt: activation.MigrationReceiptPin{ID: claim.MigrationReceiptID,
			SHA256: claim.MigrationReceiptSHA256}}
	return activation.Grant{Binding: binding, Revision: claim.Revision, LeaseUntil: claim.LeaseUntil}
}

func validActivationRequest(request ActivationRequest, completion ActivationCompletion) bool {
	return request.Plan.IsValid() && request.OperationID == request.Plan.OperationID() &&
		request.OperationID != uuid.Nil && completion.OperationID == request.OperationID &&
		completion.ReceiptID != uuid.Nil &&
		validDigest(request.LeaseTokenSHA256) && request.ExpectedRevision > 0 &&
		validDigest(completion.ReceiptSHA256) && validDigest(completion.EvidenceSHA256)
}

func (coordinator *Coordinator) RunOperatorActivation(
	ctx context.Context,
	request ActivationRequest,
	completion ActivationCompletion,
	callback OperatorActivationCallback,
) error {
	if callback == nil {
		return coordinator.Activate(ctx, request, completion, nil)
	}
	return coordinator.Activate(ctx, request, completion, func(
		transactionContext context.Context,
		scope activation.Scope,
		resources ActivationResources,
	) error {
		return callback(
			transactionContext,
			operatorActivationBinding(scope.Binding()),
			OperatorActivationResources{
				BindOwner: resources.BindOwner,
				Consume: func(command OperatorActivationCommand) error {
					return scope.Consume(activation.Command(command))
				},
			},
		)
	})
}

func operatorActivationBinding(value activation.Binding) OperatorActivationBinding {
	return OperatorActivationBinding{
		OperationID: value.OperationID, EnvironmentID: value.EnvironmentID,
		TenantID: value.TenantID, FirstUserID: value.FirstUserID,
		RootOrganizationID:      value.RootOrganizationID,
		ProtectedAdminBindingID: value.ProtectedAdminBindingID,
		InstanceOwnerBindingID:  value.InstanceOwnerBindingID,
		Kind:                    value.Kind, DeploymentMode: value.DeploymentMode,
		BindingPurpose: value.BindingPurpose, SchemaName: value.SchemaName,
		RequestSHA256: value.RequestSHA256, ManifestSHA256: value.ManifestSHA256,
		StepPlanSHA256:              value.StepPlanSHA256,
		MigrationReceiptID:          value.MigrationReceipt.ID,
		MigrationReceiptSHA256:      value.MigrationReceipt.SHA256,
		OfferingVersionID:           value.Offering.VersionID,
		OfferingRevision:            value.Offering.Revision,
		OfferingSHA256:              value.Offering.SHA256,
		OfferingEffectiveAt:         value.Offering.EffectiveAt,
		OfferingSourceSequence:      value.Offering.Source.Sequence,
		OfferingSourceSHA256:        value.Offering.Source.SHA256,
		OfferingSourcePresent:       value.Offering.Source.Present,
		RecoverySigningGeneration:   value.RecoverySigning.Generation,
		RecoveryTrustBundleSHA256:   value.RecoverySigning.TrustBundleSHA256,
		RecoverySigningPresent:      value.RecoverySigning.Present,
		AuthorityRegistryBindingID:  value.Authority.RegistryBindingID,
		AuthorityScopeKind:          value.Authority.ScopeKind,
		AuthorityScopeID:            value.Authority.ScopeID,
		AuthorityGeneration:         value.Authority.Generation,
		AuthorityPreviousHeadSHA256: value.Authority.PreviousHeadSHA256,
		AuthorityStateSHA256:        value.Authority.StateSHA256,
		AuthorityProposalSHA256:     value.Authority.ProposalSHA256,
		AuthorityAckSHA256:          value.Authority.AckSHA256,
	}
}

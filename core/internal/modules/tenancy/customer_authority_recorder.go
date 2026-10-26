package tenancy

import (
	"context"
	"fmt"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	workerapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerworker"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	"github.com/google/uuid"
)

type CustomerAuthorityRecorder struct {
	onboarding *onboardingapp.Coordinator
}

func NewCustomerAuthorityRecorder(
	onboarding *onboardingapp.Coordinator,
) (*CustomerAuthorityRecorder, error) {
	if onboarding == nil {
		return nil, tenantapi.ErrOnboardingUnavailable
	}
	return &CustomerAuthorityRecorder{onboarding: onboarding}, nil
}

func (recorder *CustomerAuthorityRecorder) RecordAuthority(
	ctx context.Context,
	claim workerapp.Claim,
	receipt workerapp.AuthorityReceipt,
) (int64, error) {
	ack, err := customerAuthorityAcknowledgement(claim, receipt)
	if recorder == nil || recorder.onboarding == nil || err != nil {
		return 0, tenantapi.ErrOnboardingInvalid
	}
	transition, err := recorder.onboarding.ReceiveAuthorityAck(ctx, ack, claim.Revision)
	if err != nil {
		return 0, fmt.Errorf("record customer authority acknowledgement: %w", err)
	}
	return transition.Revision, nil
}

func customerAuthorityAcknowledgement(
	claim workerapp.Claim,
	receipt workerapp.AuthorityReceipt,
) (tenantapi.AuthorityAcknowledgement, error) {
	plan := claim.Plan.Authority()
	operationID, operationErr := uuid.Parse(receipt.OperationID)
	bindingID, bindingErr := uuid.Parse(receipt.RegistryBindingID)
	scopeID, scopeErr := uuid.Parse(receipt.ScopeID)
	if !claim.Valid() || operationErr != nil || bindingErr != nil || scopeErr != nil ||
		tenantapi.UUID(operationID) != claim.Plan.OperationID() ||
		tenantapi.UUID(bindingID) != plan.RegistryBindingID ||
		receipt.ScopeKind != string(plan.ScopeKind) || tenantapi.UUID(scopeID) != plan.ScopeID ||
		receipt.Generation != plan.Generation || receipt.StateSHA256 != plan.StateSHA256 ||
		receipt.ProposalSHA256 != plan.ProposalSHA256 {
		return tenantapi.AuthorityAcknowledgement{}, tenantapi.ErrOnboardingConflict
	}
	ack := tenantapi.AuthorityAcknowledgement{OperationID: tenantapi.UUID(operationID),
		RegistryBindingID: tenantapi.UUID(bindingID), ScopeKind: plan.ScopeKind,
		ScopeID: tenantapi.UUID(scopeID), Generation: receipt.Generation,
		PreviousHeadSHA256: plan.PreviousHeadSHA256, StateSHA256: receipt.StateSHA256,
		ProposalSHA256: receipt.ProposalSHA256, AckSHA256: receipt.AckSHA256,
		RegistrySequence: receipt.RegistrySequence, VerifiedAt: receipt.CompletedAt}
	if !ack.IsValid() {
		return tenantapi.AuthorityAcknowledgement{}, tenantapi.ErrOnboardingInvalid
	}
	return ack, nil
}

var _ workerapp.AuthorityRecorder = (*CustomerAuthorityRecorder)(nil)

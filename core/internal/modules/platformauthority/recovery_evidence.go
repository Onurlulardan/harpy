package platformauthority

import (
	"context"

	evidencepostgres "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres/recoveryevidence"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	evidenceapp "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application/recoveryevidence"
	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports/recoveryevidence"
)

type recoveryEvidenceFacade struct {
	service *evidenceapp.Service
}

func NewRecoveryEvidence(
	database any,
) (authorityapi.RecoveryEvidence, error) {
	store, err := evidencepostgres.Bind(database)
	if err != nil {
		return nil, err
	}
	service, err := evidenceapp.New(store)
	if err != nil {
		return nil, err
	}
	return &recoveryEvidenceFacade{service: service}, nil
}

func (facade *recoveryEvidenceFacade) ReadRecoveryHead(
	ctx context.Context,
	environmentID string,
) (authorityapi.RecoveryHead, error) {
	head, err := facade.service.ReadRecoveryHead(ctx, environmentID)
	if err != nil {
		return authorityapi.RecoveryHead{}, err
	}
	return exportRecoveryHead(head), nil
}

func exportRecoveryHead(head evidenceports.Head) authorityapi.RecoveryHead {
	return authorityapi.RecoveryHead{
		BindingID: head.BindingID, EnvironmentID: head.EnvironmentID,
		DeploymentMode: head.DeploymentMode, RegistryKind: head.RegistryKind,
		LicenseBindingID:          head.LicenseBindingID,
		FinalizationReceiptSHA256: head.FinalizationReceiptSHA256,
		RecoveryTrustBundleSHA256: head.RecoveryTrustBundleSHA256,
		SignerGeneration:          head.SignerGeneration, SignerKID: head.SignerKID,
		SignerPublicKeySHA256: head.SignerPublicKeySHA256,
		BindingProofSHA256:    head.BindingProofSHA256,
		BindingProofBytes:     head.BindingProofBytes, OperationID: head.OperationID,
		Purpose: head.Purpose, ScopeKind: head.ScopeKind, ScopeID: head.ScopeID,
		Generation: head.Generation, PreviousHeadSHA256: head.PreviousHeadSHA256,
		StateSHA256: head.StateSHA256, StateBytes: head.StateBytes,
		ProposalSHA256: head.ProposalSHA256, ProposalBytes: head.ProposalBytes,
		AckSHA256: head.AckSHA256, AckBytes: head.AckBytes,
		RegistrySequence: head.RegistrySequence, CompletedAt: head.CompletedAt,
	}
}

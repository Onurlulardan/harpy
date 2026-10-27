package iam

import (
	"context"

	evidencepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/recoveryevidence"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	evidenceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/recoveryevidence"
	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryevidence"
)

type recoveryEvidenceFacade struct {
	service *evidenceapp.Service
}

func NewRecoveryEvidence(
	database any,
) (iamapi.RecoveryEvidence, error) {
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

func (facade *recoveryEvidenceFacade) ReadOwnerAnchors(
	ctx context.Context,
	environmentID string,
) ([]iamapi.RecoveryAnchor, error) {
	anchors, err := facade.service.ReadOwnerAnchors(ctx, environmentID)
	if err != nil {
		return nil, err
	}
	result := make([]iamapi.RecoveryAnchor, len(anchors))
	for index, anchor := range anchors {
		result[index] = exportRecoveryAnchor(anchor)
	}
	return result, nil
}

func exportRecoveryAnchor(anchor evidenceports.Anchor) iamapi.RecoveryAnchor {
	return iamapi.RecoveryAnchor{
		EnvironmentID: anchor.EnvironmentID, TenantID: anchor.TenantID,
		OwnerUserID: anchor.OwnerUserID, OwnerBindingID: anchor.OwnerBindingID,
		OwnerBindingGeneration: anchor.OwnerBindingGeneration,
		AnchorID:               anchor.AnchorID, AnchorGeneration: anchor.AnchorGeneration,
		CredentialID: anchor.CredentialID, PublicKeyCOSE: anchor.PublicKeyCOSE,
		AAGUID: anchor.AAGUID, AttestationFormat: anchor.AttestationFormat,
		AttestationType: anchor.AttestationType, RPID: anchor.RPID,
		AllowedOrigins:       anchor.AllowedOrigins,
		AllowedOriginsSHA256: anchor.AllowedOriginsSHA256,
		UserVerification:     anchor.UserVerification,
		BackupEligible:       anchor.BackupEligible, BackupState: anchor.BackupState,
		SignCount: anchor.SignCount, CounterPolicy: anchor.CounterPolicy,
		ProtectedHeadGeneration: anchor.ProtectedHeadGeneration,
		ProtectedStateSHA256:    anchor.ProtectedStateSHA256,
		ProtectedAckSHA256:      anchor.ProtectedAckSHA256,
	}
}

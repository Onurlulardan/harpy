package iam

import (
	"context"
	"errors"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	invitationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type systemTransferInvitationRoleResolver interface {
	FindInvitationRole(context.Context, platformtx.TenantScope,
		publicresource.ActorEvidence, publicresource.ID,
	) (invitationapp.RoleScope, error)
}

type systemTransferInvitationCommand struct {
	application *invitationHTTPApplication
	roles       systemTransferInvitationRoleResolver
}

func newSystemTransferInvitationCommand(application *invitationHTTPApplication,
	roles systemTransferInvitationRoleResolver,
) (*systemTransferInvitationCommand, error) {
	if application == nil || application.runtime == nil ||
		!application.runtime.Valid() || application.sealer == nil ||
		nilIdentityDependency(roles) {
		return nil, systemtransfer.ErrDependencies
	}
	return &systemTransferInvitationCommand{application: application, roles: roles}, nil
}

func (command *systemTransferInvitationCommand) Invite(ctx context.Context,
	scope platformtx.TenantScope, evidence publicresource.ActorEvidence,
	input systemtransfer.InvitationInput,
) (systemtransfer.InvitationResult, error) {
	if command == nil || input.Validate() != nil ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationInvalid
	}
	roleScope, err := command.roles.FindInvitationRole(
		ctx, scope, evidence, input.RoleID)
	if err != nil {
		return systemtransfer.InvitationResult{}, err
	}
	email, err := (invitationNormalizer{}).NormalizeEmail(input.Email)
	if err != nil {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationInvalid
	}
	tokenVersion, authorizationVersion, err := evidence.Versions()
	if err != nil || tokenVersion < 1 {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationDenied
	}
	operationID := uuid.NewSHA1(uuid.NameSpaceOID,
		[]byte("harpy/iam/system-import/invitation/v1/"+input.IdempotencyKey))
	invitationID := invitationHTTPChildID(operationID, "invitation")
	tokenID := invitationHTTPChildID(operationID, "token")
	artifactID := invitationHTTPChildID(operationID, "artifact")
	requestSHA, err := identityRequestSHA(struct {
		Email  string
		RoleID uuid.UUID
		Scope  invitationapp.RoleScope
	}{email.NormalizedValue, uuid.UUID(input.RoleID), roleScope})
	if err != nil {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationInvalid
	}
	issue, token, err := invitationapp.NewCredentialGenerator().GenerateCredential(
		scope.TenantID(), command.application.keys)
	if err != nil {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationUnavailable
	}
	defer issue.Destroy()
	var envelope deliveryartifactv1.Envelope
	err = issue.RevealOnce(func(plaintext []byte) error {
		var sealErr error
		envelope, sealErr = command.application.sealer.SealInvitationCredential(
			ctx, scope.TenantID(), artifactID, invitationID, tokenID, plaintext)
		return sealErr
	})
	reference, referenceErr := deliveryartifactv1.NewReference(scope.TenantID(), artifactID)
	if err != nil || referenceErr != nil || envelope.Validate(scope.TenantID()) != nil {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationUnavailable
	}
	created, err := command.application.runtime.service.Create(ctx, scope,
		invitationapp.CreateCommand{
			OperationID: operationID, RequestSHA256: requestSHA, ID: invitationID,
			DisplayValue: input.Email, NormalizedValue: email.NormalizedValue,
			NormalizedSHA256: email.SHA256, RoleID: uuid.UUID(input.RoleID),
			RoleScope: roleScope, InvitedBy: scope.ActorID(),
			ExpectedAuthorizationVersion: authorizationVersion,
			Token: invitationapp.TokenIssue{ID: tokenID, Selector: token.Selector,
				HashKeyID: token.HashKeyID, SecretHMAC: token.SecretHMAC,
				BindingSHA256:       token.BindingSHA256,
				DeliveryArtifactRef: reference.String(), DeliveryEnvelope: envelope},
		})
	if err != nil {
		return systemtransfer.InvitationResult{}, mapSystemTransferInvitationError(err)
	}
	result := systemtransfer.InvitationResult{ID: publicresource.ID(created.ID),
		Revision: created.Revision, Replayed: created.Replayed}
	if created.Validate() != nil || result.Validate() != nil {
		return systemtransfer.InvitationResult{}, systemtransfer.ErrInvitationUnavailable
	}
	return result, nil
}

func mapSystemTransferInvitationError(err error) error {
	switch {
	case errors.Is(err, invitationapp.ErrInvalid):
		return systemtransfer.ErrInvitationInvalid
	case errors.Is(err, invitationapp.ErrDenied):
		return systemtransfer.ErrInvitationDenied
	case errors.Is(err, invitationapp.ErrConflict), errors.Is(err, invitationapp.ErrStale):
		return systemtransfer.ErrInvitationConflict
	default:
		return systemtransfer.ErrInvitationUnavailable
	}
}

var _ systemtransfer.InvitationCommand = (*systemTransferInvitationCommand)(nil)

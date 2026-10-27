package iam

import (
	"context"
	"strings"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *invitationHTTPApplication) Create(
	ctx context.Context,
	claims accesstoken.Claims,
	input InvitationCreateInput,
) (InvitationView, error) {
	input.Email, input.Role = strings.TrimSpace(input.Email), strings.TrimSpace(input.Role)
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validInvitationClaims(claims) ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) ||
		input.Email == "" || input.Role == "" {
		return InvitationView{}, ErrInvitationHTTPInvalid
	}
	evidence, err := (invitationNormalizer{}).NormalizeEmail(input.Email)
	if err != nil {
		return InvitationView{}, ErrInvitationHTTPInvalid
	}
	role, err := application.invitationRole(ctx, claims, input.Role)
	if err != nil {
		return InvitationView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.invitation.create", input.IdempotencyKey,
	)
	invitationID := invitationHTTPChildID(operationID, "invitation")
	tokenID := invitationHTTPChildID(operationID, "token")
	artifactID := invitationHTTPChildID(operationID, "artifact")
	requestSHA, err := identityRequestSHA(struct {
		Email, Role string
		RoleID      uuid.UUID
		Scope       invitation.RoleScope
	}{evidence.NormalizedValue, role.Name, role.ID, role.Scope})
	if err != nil {
		return InvitationView{}, ErrInvitationHTTPInvalid
	}
	issued, tokenEvidence, err := invitation.NewCredentialGenerator().GenerateCredential(
		claims.Tenant(), application.keys,
	)
	if err != nil {
		return InvitationView{}, ErrInvitationHTTPUnavailable
	}
	defer issued.Destroy()
	var envelope deliveryartifactv1.Envelope
	err = issued.RevealOnce(func(plaintext []byte) error {
		var sealErr error
		envelope, sealErr = application.sealer.SealInvitationCredential(
			ctx, claims.Tenant(), artifactID, invitationID, tokenID, plaintext,
		)
		return sealErr
	})
	if err != nil || envelope.Validate(claims.Tenant()) != nil {
		return InvitationView{}, ErrInvitationHTTPUnavailable
	}
	reference, err := deliveryartifactv1.NewReference(claims.Tenant(), artifactID)
	if err != nil {
		return InvitationView{}, ErrInvitationHTTPUnavailable
	}
	command := invitation.CreateCommand{
		OperationID: operationID, RequestSHA256: requestSHA, ID: invitationID,
		DisplayValue: input.Email, NormalizedValue: evidence.NormalizedValue,
		NormalizedSHA256: evidence.SHA256, RoleID: role.ID, RoleScope: role.Scope,
		InvitedBy:                    claims.Subject(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
		Token: invitation.TokenIssue{
			ID: tokenID, Selector: tokenEvidence.Selector,
			HashKeyID: tokenEvidence.HashKeyID, SecretHMAC: tokenEvidence.SecretHMAC,
			BindingSHA256:       tokenEvidence.BindingSHA256,
			DeliveryArtifactRef: reference.String(), DeliveryEnvelope: envelope,
		},
	}
	request, err := application.commandScope(ctx, claims, "iam.invitation.create")
	if err != nil {
		return InvitationView{}, err
	}
	err = application.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			_, createErr := application.runtime.service.Create(scoped, scope, command)
			return createErr
		},
	)
	if err != nil {
		return InvitationView{}, mapInvitationApplicationError(err)
	}
	return application.Read(ctx, claims, invitationID)
}

func invitationHTTPChildID(parent uuid.UUID, purpose string) uuid.UUID {
	return uuid.NewSHA1(parent, []byte("harpy/iam/invitation/"+purpose+"/v1"))
}

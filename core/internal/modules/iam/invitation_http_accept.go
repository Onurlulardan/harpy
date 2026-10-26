package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"github.com/google/uuid"
)

func (application *invitationHTTPApplication) Accept(
	ctx context.Context,
	input InvitationAcceptInput,
) (InvitationAcceptView, error) {
	defer input.Destroy()
	if application == nil || ctx == nil || ctx.Err() != nil ||
		input.TenantSlug == "" || input.FirstName == "" || input.LastName == "" ||
		len(input.Token) == 0 || len(input.Password) == 0 {
		return InvitationAcceptView{}, ErrInvitationHTTPInvalid
	}
	presented, err := parseInvitationCredential(input.Token)
	if err != nil {
		return InvitationAcceptView{}, err
	}
	defer presented.Destroy()
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return InvitationAcceptView{}, ErrInvitationHTTPInvalidCredential
	}
	record, _, err := application.invitationCredential(ctx, tenantID, presented)
	if err != nil || (record.Invitation.State == invitation.StatePending &&
		!time.Now().UTC().Before(record.Invitation.ExpiresAt)) {
		if errors.Is(err, ErrInvitationHTTPUnavailable) {
			return InvitationAcceptView{}, err
		}
		return InvitationAcceptView{}, ErrInvitationHTTPInvalidCredential
	}
	credential, passwordSHA, err := application.invitationPassword(ctx, input, record)
	if err != nil {
		return InvitationAcceptView{}, err
	}
	directoryHMAC := invitationDirectoryHMAC(
		application.directory.material, record.Invitation.NormalizedValue,
	)
	requestSHA, err := identityRequestSHA(struct {
		InvitationID uuid.UUID
		First, Last  string
		Custom       []byte
		Password     [32]byte
	}{record.Invitation.ID, input.FirstName, input.LastName,
		input.CustomAttributes, passwordSHA})
	if err != nil {
		clear(directoryHMAC[:])
		return InvitationAcceptView{}, ErrInvitationHTTPInvalid
	}
	selector, _ := presented.Selector()
	operationID := identityDeterministicUUID(
		[]byte("harpy/iam/invitation-accept/v1"), tenantID[:],
		record.Invitation.ID[:], selector[:], requestSHA[:],
	)
	ids := invitationRegistrationIDs(operationID, record.Invitation.RoleScope.Kind)
	err = application.acceptInvitation(
		ctx, tenantID, presented, record, input, credential,
		directoryHMAC, operationID, requestSHA, ids,
	)
	clear(directoryHMAC[:])
	if err != nil {
		return InvitationAcceptView{}, err
	}
	return InvitationAcceptView{
		UserID: ids.user, TenantSlug: input.TenantSlug, RequiresLogin: true,
	}, nil
}

func (application *invitationHTTPApplication) invitationPassword(
	ctx context.Context,
	input InvitationAcceptInput,
	record invitation.CredentialRecord,
) (password.Argon2idCredential, [32]byte, error) {
	secret, err := application.passwords.NewSecret(ctx, input.Password, []string{
		record.Invitation.DisplayValue, record.Invitation.NormalizedValue,
		input.FirstName, input.LastName,
	})
	if err != nil {
		return password.Argon2idCredential{}, [32]byte{}, ErrInvitationHTTPInvalid
	}
	defer secret.Destroy()
	credential, err := application.hasher.Hash(secret)
	if err != nil {
		return password.Argon2idCredential{}, [32]byte{}, ErrInvitationHTTPUnavailable
	}
	return credential, sha256.Sum256(input.Password), nil
}

func invitationDirectoryHMAC(key [32]byte, normalized string) [32]byte {
	digest := hmac.New(sha256.New, key[:])
	_, _ = digest.Write([]byte(normalized))
	value := digest.Sum(nil)
	var result [32]byte
	copy(result[:], value)
	clear(value)
	return result
}

type invitationRegistrationIDSet struct {
	user, identifier, accessMode, assignment, membership uuid.UUID
}

func invitationRegistrationIDs(
	operationID uuid.UUID, kind invitation.RoleScopeKind,
) invitationRegistrationIDSet {
	ids := invitationRegistrationIDSet{
		user:       invitationHTTPChildID(operationID, "accepted-user"),
		identifier: invitationHTTPChildID(operationID, "accepted-email"),
		accessMode: invitationHTTPChildID(operationID, "operator-access"),
		assignment: invitationHTTPChildID(operationID, "role-assignment"),
	}
	if kind == invitation.RoleScopeOrganization {
		ids.membership = invitationHTTPChildID(operationID, "membership")
	}
	return ids
}

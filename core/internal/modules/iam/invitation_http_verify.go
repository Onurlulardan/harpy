package iam

import (
	"context"
	"errors"
	"time"
)

func (application *invitationHTTPApplication) Verify(
	ctx context.Context,
	input InvitationCredentialInput,
) (InvitationVerifyView, error) {
	defer clear(input.Token)
	if application == nil || ctx == nil || ctx.Err() != nil ||
		input.TenantSlug == "" || len(input.Token) == 0 {
		return InvitationVerifyView{}, ErrInvitationHTTPInvalidCredential
	}
	presented, err := parseInvitationCredential(input.Token)
	if err != nil {
		return InvitationVerifyView{}, err
	}
	defer presented.Destroy()
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return InvitationVerifyView{}, ErrInvitationHTTPInvalidCredential
	}
	record, _, err := application.invitationCredential(ctx, tenantID, presented)
	if err != nil || record.Invitation.State != "pending" ||
		!time.Now().UTC().Before(record.Invitation.ExpiresAt) {
		if errors.Is(err, ErrInvitationHTTPUnavailable) {
			return InvitationVerifyView{}, err
		}
		return InvitationVerifyView{}, ErrInvitationHTTPInvalidCredential
	}
	return InvitationVerifyView{
		Email: record.Invitation.DisplayValue, Role: record.Invitation.RoleName,
		TenantName: input.TenantSlug, TenantSlug: input.TenantSlug,
		ExpiresAt: record.Invitation.ExpiresAt,
	}, nil
}

package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *invitationHTTPApplication) List(
	ctx context.Context,
	claims accesstoken.Claims,
	input InvitationListInput,
) (InvitationPageView, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validInvitationClaims(claims) || input.Limit < 1 || input.Limit > 100 {
		return InvitationPageView{}, ErrInvitationHTTPInvalid
	}
	states := make([]invitation.State, 0, len(input.States))
	for _, raw := range input.States {
		state := invitation.State(strings.ToLower(raw))
		if !state.Valid() {
			return InvitationPageView{}, ErrInvitationHTTPInvalid
		}
		states = append(states, state)
	}
	request, err := application.snapshotScope(ctx, claims)
	if err != nil {
		return InvitationPageView{}, err
	}
	var page invitation.Page
	err = application.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			page, listErr = application.runtime.service.List(
				scoped, scope, invitation.ListQuery{
					States: states, Limit: input.Limit, After: input.After,
				},
			)
			return listErr
		},
	)
	if err != nil {
		return InvitationPageView{}, mapInvitationApplicationError(err)
	}
	view := InvitationPageView{
		Items: make([]InvitationView, 0, len(page.Invitations)), Next: page.Next,
	}
	for _, record := range page.Invitations {
		view.Items = append(view.Items, invitationHTTPView(record))
	}
	return view, nil
}

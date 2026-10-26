package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

type portalUserHTTPTestCall struct {
	operation string
	input     PortalUserMutationInput
	create    PortalUserCreateInput
}

type portalUserHTTPTestApplication struct {
	calls []portalUserHTTPTestCall
	view  PortalUserView
	err   error
}

func (application *portalUserHTTPTestApplication) CreatePortalUser(
	_ context.Context, _ accesstoken.Claims, input PortalUserCreateInput,
) (PortalUserView, error) {
	application.calls = append(application.calls,
		portalUserHTTPTestCall{operation: "create", create: input})
	return application.view, application.err
}

func (application *portalUserHTTPTestApplication) InvitePortalUser(
	_ context.Context, _ accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	return application.mutate("invite", input)
}

func (application *portalUserHTTPTestApplication) ChangePortalUserStatus(
	_ context.Context, _ accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	return application.mutate("status", input)
}

func (application *portalUserHTTPTestApplication) GrantPortalConsole(
	_ context.Context, _ accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	return application.mutate("grant", input)
}

func (application *portalUserHTTPTestApplication) RevokePortalConsole(
	_ context.Context, _ accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	return application.mutate("revoke", input)
}

func (application *portalUserHTTPTestApplication) mutate(
	operation string, input PortalUserMutationInput,
) (PortalUserView, error) {
	application.calls = append(application.calls,
		portalUserHTTPTestCall{operation: operation, input: input})
	return application.view, application.err
}

package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

type invitationHTTPTestApplication struct {
	createInput  InvitationCreateInput
	verifyInput  InvitationCredentialInput
	acceptInput  InvitationAcceptInput
	createResult InvitationView
	verifyResult InvitationVerifyView
	acceptResult InvitationAcceptView
	err          error
	createCalls  int
	listCalls    int
	verifyCalls  int
	acceptCalls  int
}

func (app *invitationHTTPTestApplication) Create(
	_ context.Context, _ accesstoken.Claims, input InvitationCreateInput,
) (InvitationView, error) {
	app.createCalls++
	app.createInput = input
	return app.createResult, app.err
}

func (app *invitationHTTPTestApplication) Read(
	context.Context, accesstoken.Claims, uuid.UUID,
) (InvitationView, error) {
	return app.createResult, app.err
}

func (app *invitationHTTPTestApplication) List(
	context.Context, accesstoken.Claims, InvitationListInput,
) (InvitationPageView, error) {
	app.listCalls++
	return InvitationPageView{}, app.err
}

func (app *invitationHTTPTestApplication) Revoke(
	context.Context, accesstoken.Claims, InvitationRevokeInput,
) error {
	return app.err
}

func (app *invitationHTTPTestApplication) Verify(
	_ context.Context, input InvitationCredentialInput,
) (InvitationVerifyView, error) {
	app.verifyCalls++
	app.verifyInput = InvitationCredentialInput{
		Token: append([]byte(nil), input.Token...), TenantSlug: input.TenantSlug,
	}
	return app.verifyResult, app.err
}

func (app *invitationHTTPTestApplication) Accept(
	_ context.Context, input InvitationAcceptInput,
) (InvitationAcceptView, error) {
	app.acceptCalls++
	app.acceptInput = InvitationAcceptInput{
		Token: append([]byte(nil), input.Token...), TenantSlug: input.TenantSlug,
		FirstName: input.FirstName, LastName: input.LastName,
		Password:         append([]byte(nil), input.Password...),
		CustomAttributes: append([]byte(nil), input.CustomAttributes...),
	}
	return app.acceptResult, app.err
}

package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/google/uuid"
)

func (application *identityHTTPTestApplication) SystemResources() publicresource.Port {
	return nil
}

type identityHTTPTestCall struct {
	operation string
	claims    accesstoken.Claims
	input     any
}

type identityHTTPTestQuery struct {
	resource IdentityResource
	request  queryapp.Request
}

type identityHTTPTestApplication struct {
	calls    []identityHTTPTestCall
	results  map[string]any
	failures map[string]error
}

func newIdentityHTTPTestApplication() *identityHTTPTestApplication {
	return &identityHTTPTestApplication{
		results: make(map[string]any), failures: make(map[string]error),
	}
}

func (application *identityHTTPTestApplication) invoke(
	operation string, claims accesstoken.Claims, input any,
) (any, error) {
	application.calls = append(application.calls, identityHTTPTestCall{
		operation: operation, claims: claims, input: input,
	})
	return application.results[operation], application.failures[operation]
}

func (application *identityHTTPTestApplication) CreateUser(
	_ context.Context, claims accesstoken.Claims, input IdentityUserCreate,
) (IdentityUserView, error) {
	result, err := application.invoke("CreateUser", claims, input)
	view, _ := result.(IdentityUserView)
	return view, err
}

func (application *identityHTTPTestApplication) GetUser(
	_ context.Context, claims accesstoken.Claims, id uuid.UUID,
) (IdentityUserView, error) {
	result, err := application.invoke("GetUser", claims, id)
	view, _ := result.(IdentityUserView)
	return view, err
}

func (application *identityHTTPTestApplication) UpdateUser(
	_ context.Context, claims accesstoken.Claims, input IdentityUserUpdate,
) (IdentityUserView, error) {
	result, err := application.invoke("UpdateUser", claims, input)
	view, _ := result.(IdentityUserView)
	return view, err
}

func (application *identityHTTPTestApplication) DeleteUser(
	_ context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	_, err := application.invoke("DeleteUser", claims, input)
	return err
}

func (application *identityHTTPTestApplication) ChangeUserState(
	_ context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityUserView, error) {
	result, err := application.invoke("ChangeUserState", claims, input)
	view, _ := result.(IdentityUserView)
	return view, err
}

func (application *identityHTTPTestApplication) CreateRole(
	_ context.Context, claims accesstoken.Claims, input IdentityRoleCreate,
) (IdentityRoleView, error) {
	result, err := application.invoke("CreateRole", claims, input)
	view, _ := result.(IdentityRoleView)
	return view, err
}

func (application *identityHTTPTestApplication) GetRole(
	_ context.Context, claims accesstoken.Claims, id uuid.UUID,
) (IdentityRoleView, error) {
	result, err := application.invoke("GetRole", claims, id)
	view, _ := result.(IdentityRoleView)
	return view, err
}

func (application *identityHTTPTestApplication) UpdateRole(
	_ context.Context, claims accesstoken.Claims, input IdentityRoleUpdate,
) (IdentityRoleView, error) {
	result, err := application.invoke("UpdateRole", claims, input)
	view, _ := result.(IdentityRoleView)
	return view, err
}

func (application *identityHTTPTestApplication) DeleteRole(
	_ context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	_, err := application.invoke("DeleteRole", claims, input)
	return err
}

func (application *identityHTTPTestApplication) ChangeRoleState(
	_ context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityRoleView, error) {
	result, err := application.invoke("ChangeRoleState", claims, input)
	view, _ := result.(IdentityRoleView)
	return view, err
}

func (application *identityHTTPTestApplication) CreateOrganization(
	_ context.Context, claims accesstoken.Claims, input IdentityOrganizationCreate,
) (IdentityOrganizationView, error) {
	result, err := application.invoke("CreateOrganization", claims, input)
	view, _ := result.(IdentityOrganizationView)
	return view, err
}

func (application *identityHTTPTestApplication) GetOrganization(
	_ context.Context, claims accesstoken.Claims, id uuid.UUID,
) (IdentityOrganizationView, error) {
	result, err := application.invoke("GetOrganization", claims, id)
	view, _ := result.(IdentityOrganizationView)
	return view, err
}

func (application *identityHTTPTestApplication) UpdateOrganization(
	_ context.Context, claims accesstoken.Claims, input IdentityOrganizationUpdate,
) (IdentityOrganizationView, error) {
	result, err := application.invoke("UpdateOrganization", claims, input)
	view, _ := result.(IdentityOrganizationView)
	return view, err
}

func (application *identityHTTPTestApplication) DeleteOrganization(
	_ context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	_, err := application.invoke("DeleteOrganization", claims, input)
	return err
}

func (application *identityHTTPTestApplication) ChangeOrganizationState(
	_ context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityOrganizationView, error) {
	result, err := application.invoke("ChangeOrganizationState", claims, input)
	view, _ := result.(IdentityOrganizationView)
	return view, err
}

func (application *identityHTTPTestApplication) Query(
	_ context.Context, claims accesstoken.Claims,
	resource IdentityResource, request queryapp.Request,
) (IdentityQueryView, error) {
	result, err := application.invoke(
		"Query", claims, identityHTTPTestQuery{resource: resource, request: request},
	)
	view, _ := result.(IdentityQueryView)
	return view, err
}

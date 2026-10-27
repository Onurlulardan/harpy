package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (application *identityHTTPTestApplication) DeleteUsers(
	_ context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	result, err := application.invoke("DeleteUsers", claims, input)
	count, _ := result.(int)
	return count, err
}

func (application *identityHTTPTestApplication) DeleteRoles(
	_ context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	result, err := application.invoke("DeleteRoles", claims, input)
	count, _ := result.(int)
	return count, err
}

func (application *identityHTTPTestApplication) DeleteOrganizations(
	_ context.Context, claims accesstoken.Claims, input IdentityBulkDelete,
) (int, error) {
	result, err := application.invoke("DeleteOrganizations", claims, input)
	count, _ := result.(int)
	return count, err
}

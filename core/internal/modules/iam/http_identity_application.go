package iam

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/google/uuid"
)

var (
	ErrIdentityHTTPForbidden        = errors.New("IAM system resource operation is forbidden")
	ErrIdentityHTTPNotFound         = errors.New("IAM system resource was not found")
	ErrIdentityHTTPConflict         = errors.New("IAM system resource operation conflicts")
	ErrIdentityHTTPInvalid          = errors.New("IAM system resource operation is invalid")
	ErrIdentityHTTPRecoveryRequired = errors.New("IAM profile recovery is required")
	ErrIdentityHTTPUnavailable      = errors.New("IAM system resource operation is unavailable")
)

type IdentityHTTPApplication interface {
	SystemResources() publicresource.Port
	CreateUser(context.Context, accesstoken.Claims, IdentityUserCreate) (IdentityUserView, error)
	GetUser(context.Context, accesstoken.Claims, uuid.UUID) (IdentityUserView, error)
	UpdateUser(context.Context, accesstoken.Claims, IdentityUserUpdate) (IdentityUserView, error)
	DeleteUser(context.Context, accesstoken.Claims, IdentityMutation) error
	DeleteUsers(context.Context, accesstoken.Claims, IdentityBulkDelete) (int, error)
	ChangeUserState(context.Context, accesstoken.Claims, IdentityStateMutation) (IdentityUserView, error)

	CreateRole(context.Context, accesstoken.Claims, IdentityRoleCreate) (IdentityRoleView, error)
	GetRole(context.Context, accesstoken.Claims, uuid.UUID) (IdentityRoleView, error)
	UpdateRole(context.Context, accesstoken.Claims, IdentityRoleUpdate) (IdentityRoleView, error)
	DeleteRole(context.Context, accesstoken.Claims, IdentityMutation) error
	DeleteRoles(context.Context, accesstoken.Claims, IdentityBulkDelete) (int, error)
	ChangeRoleState(context.Context, accesstoken.Claims, IdentityStateMutation) (IdentityRoleView, error)

	CreateOrganization(context.Context, accesstoken.Claims, IdentityOrganizationCreate) (IdentityOrganizationView, error)
	GetOrganization(context.Context, accesstoken.Claims, uuid.UUID) (IdentityOrganizationView, error)
	UpdateOrganization(context.Context, accesstoken.Claims, IdentityOrganizationUpdate) (IdentityOrganizationView, error)
	DeleteOrganization(context.Context, accesstoken.Claims, IdentityMutation) error
	DeleteOrganizations(context.Context, accesstoken.Claims, IdentityBulkDelete) (int, error)
	ChangeOrganizationState(context.Context, accesstoken.Claims, IdentityStateMutation) (IdentityOrganizationView, error)

	Query(context.Context, accesstoken.Claims, IdentityResource, queryapp.Request) (IdentityQueryView, error)
}

package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

var (
	ErrRBACHTTPForbidden   = errors.New("IAM RBAC operation is forbidden")
	ErrRBACHTTPNotFound    = errors.New("IAM RBAC record was not found")
	ErrRBACHTTPConflict    = errors.New("IAM RBAC operation conflicts")
	ErrRBACHTTPInvalid     = errors.New("IAM RBAC operation is invalid")
	ErrRBACHTTPUnavailable = errors.New("IAM RBAC operation is unavailable")
)

type RBACHTTPApplication interface {
	CreatePermission(context.Context, accesstoken.Claims, RBACPermissionCreate) (RBACPermissionView, error)
	GetPermission(context.Context, accesstoken.Claims, uuid.UUID) (RBACPermissionView, error)
	ListPermissions(context.Context, accesstoken.Claims, RBACList) (RBACPermissionPage, error)
	UpdatePermission(context.Context, accesstoken.Claims, RBACPermissionUpdate) (RBACPermissionView, error)
	DeletePermission(context.Context, accesstoken.Claims, RBACMutation) error

	CreateResource(context.Context, accesstoken.Claims, RBACCatalogCreate) (RBACResourceView, error)
	GetResource(context.Context, accesstoken.Claims, uuid.UUID) (RBACResourceView, error)
	ListResources(context.Context, accesstoken.Claims, RBACList) (RBACResourcePage, error)
	UpdateResource(context.Context, accesstoken.Claims, RBACCatalogUpdate) (RBACResourceView, error)
	DeleteResource(context.Context, accesstoken.Claims, RBACMutation) error

	CreateAction(context.Context, accesstoken.Claims, RBACCatalogCreate) (RBACActionView, error)
	GetAction(context.Context, accesstoken.Claims, uuid.UUID) (RBACActionView, error)
	ListActions(context.Context, accesstoken.Claims, RBACList) (RBACActionPage, error)
	UpdateAction(context.Context, accesstoken.Claims, RBACCatalogUpdate) (RBACActionView, error)
	DeleteAction(context.Context, accesstoken.Claims, RBACMutation) error

	Check(context.Context, accesstoken.Claims, RBACCheck) (RBACDecisionView, error)
	BatchCheck(context.Context, accesstoken.Claims, RBACBatchCheck) (RBACBatchDecisionView, error)
	Effective(context.Context, accesstoken.Claims, RBACEffectiveRead) (RBACEffectivePage, error)
}

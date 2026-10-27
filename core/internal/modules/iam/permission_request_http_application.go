package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

var (
	ErrPermissionRequestHTTPInvalid     = errors.New("IAM permission request operation is invalid")
	ErrPermissionRequestHTTPForbidden   = errors.New("IAM permission request operation is forbidden")
	ErrPermissionRequestHTTPNotFound    = errors.New("IAM permission request was not found")
	ErrPermissionRequestHTTPConflict    = errors.New("IAM permission request operation conflicts")
	ErrPermissionRequestHTTPUnavailable = errors.New("IAM permission request operation is unavailable")
)

type PermissionRequestHTTPApplication interface {
	Create(context.Context, accesstoken.Claims, PermissionRequestCreate) (PermissionRequestRefView, error)
	List(context.Context, accesstoken.Claims, PermissionRequestList) (PermissionRequestPageView, error)
	ListCatalogResources(context.Context, accesstoken.Claims, PermissionRequestCatalogList) (PermissionRequestCatalogResourcePageView, error)
	ListCatalogActions(context.Context, accesstoken.Claims, PermissionRequestCatalogList) (PermissionRequestCatalogActionPageView, error)
	ListMissingCatalogResources(context.Context, accesstoken.Claims, PermissionRequestMissingCatalogList) (PermissionRequestCatalogResourcePageView, error)
	Get(context.Context, accesstoken.Claims, uuid.UUID) (PermissionRequestGroupView, error)
	Review(context.Context, accesstoken.Claims, PermissionRequestReview) (PermissionRequestRefView, error)
	Approve(context.Context, accesstoken.Claims, PermissionRequestDecision) (PermissionRequestRefView, error)
	Reject(context.Context, accesstoken.Claims, PermissionRequestDecision) (PermissionRequestRefView, error)
	Cancel(context.Context, accesstoken.Claims, PermissionRequestTerminal) (PermissionRequestRefView, error)
	Expire(context.Context, accesstoken.Claims, PermissionRequestTerminal) (PermissionRequestRefView, error)
}

type PermissionRequestHTTPPlatform = IdentityHTTPPlatform

package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

var (
	ErrPortalUserHTTPInvalid     = errors.New("IAM portal user request is invalid")
	ErrPortalUserHTTPForbidden   = errors.New("IAM portal user operation is forbidden")
	ErrPortalUserHTTPNotFound    = errors.New("IAM portal user was not found")
	ErrPortalUserHTTPConflict    = errors.New("IAM portal user operation conflicts")
	ErrPortalUserHTTPUnavailable = errors.New("IAM portal user operation is unavailable")
)

type PortalUserCreateInput struct {
	IdempotencyKey, Email, FirstName, LastName string
	FormSlug, SubjectType, Status              string
	SubjectID                                  uuid.UUID
}

type PortalUserMutationInput struct {
	ID               uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
	Status           string
}

type PortalUserView struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Status        string    `json:"status"`
	AccountType   string    `json:"account_type"`
	BindingID     uuid.UUID `json:"portal_binding_id"`
	SubjectType   string    `json:"portal_subject_type"`
	SubjectID     uuid.UUID `json:"portal_subject_id"`
	Revision      int64     `json:"revision"`
	SetupRequired bool      `json:"portal_setup_required"`
	InviteSent    bool      `json:"invite_sent,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PortalUserHTTPApplication interface {
	CreatePortalUser(context.Context, accesstoken.Claims, PortalUserCreateInput) (PortalUserView, error)
	InvitePortalUser(context.Context, accesstoken.Claims, PortalUserMutationInput) (PortalUserView, error)
	ChangePortalUserStatus(context.Context, accesstoken.Claims, PortalUserMutationInput) (PortalUserView, error)
	GrantPortalConsole(context.Context, accesstoken.Claims, PortalUserMutationInput) (PortalUserView, error)
	RevokePortalConsole(context.Context, accesstoken.Claims, PortalUserMutationInput) (PortalUserView, error)
}

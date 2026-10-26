package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

var (
	ErrInvitationHTTPInvalid           = errors.New("IAM invitation request is invalid")
	ErrInvitationHTTPInvalidCredential = errors.New("IAM invitation credential is invalid")
	ErrInvitationHTTPForbidden         = errors.New("IAM invitation operation is forbidden")
	ErrInvitationHTTPNotFound          = errors.New("IAM invitation was not found")
	ErrInvitationHTTPConflict          = errors.New("IAM invitation operation conflicts")
	ErrInvitationHTTPRecoveryRequired  = errors.New("IAM profile recovery is required")
	ErrInvitationHTTPUnavailable       = errors.New("IAM invitation runtime is unavailable")
)

type InvitationHTTPApplication interface {
	Create(context.Context, accesstoken.Claims, InvitationCreateInput) (InvitationView, error)
	Read(context.Context, accesstoken.Claims, uuid.UUID) (InvitationView, error)
	List(context.Context, accesstoken.Claims, InvitationListInput) (InvitationPageView, error)
	Revoke(context.Context, accesstoken.Claims, InvitationRevokeInput) error
	Verify(context.Context, InvitationCredentialInput) (InvitationVerifyView, error)
	Accept(context.Context, InvitationAcceptInput) (InvitationAcceptView, error)
}

type InvitationCreateInput struct {
	IdempotencyKey string
	Email          string
	Role           string
}

type InvitationListInput struct {
	Limit  int
	After  uuid.UUID
	States []string
}

type InvitationRevokeInput struct {
	ID               uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
}

type InvitationCredentialInput struct {
	Token      []byte
	TenantSlug string
}

type InvitationAcceptInput struct {
	Token            []byte
	TenantSlug       string
	FirstName        string
	LastName         string
	Password         []byte
	CustomAttributes []byte
}

func (input *InvitationAcceptInput) Destroy() {
	if input == nil {
		return
	}
	clear(input.Token)
	clear(input.Password)
	input.Token, input.Password = nil, nil
}

type InvitationView struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	InvitedBy uuid.UUID `json:"invited_by"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revision  int64     `json:"revision"`
}

type InvitationPageView struct {
	Items []InvitationView
	Next  uuid.UUID
}

type InvitationVerifyView struct {
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	TenantName string    `json:"tenant_name"`
	TenantSlug string    `json:"tenant_slug"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type InvitationAcceptView struct {
	UserID        uuid.UUID `json:"user_id"`
	TenantSlug    string    `json:"tenant_slug"`
	RequiresLogin bool      `json:"requires_login"`
}

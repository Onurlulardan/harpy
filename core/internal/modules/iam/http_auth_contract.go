package iam

import (
	"context"
	"errors"
	"time"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
	"github.com/google/uuid"
)

var (
	ErrAuthHTTPInvalidCredentials = errors.New("IAM browser credential is invalid")
	ErrAuthHTTPInvalidSession     = errors.New("IAM browser session is invalid")
	ErrAuthHTTPPasswordRejected   = errors.New("IAM browser password is rejected")
	ErrAuthHTTPUnavailable        = errors.New("IAM browser authentication is unavailable")
)

type AuthHTTPLoginMode string

const (
	AuthHTTPLoginOperator AuthHTTPLoginMode = "operator_console"
	AuthHTTPLoginPortal   AuthHTTPLoginMode = "portal"
)

type AuthHTTPLoginInput struct {
	Mode       AuthHTTPLoginMode
	TenantSlug string
	Identifier string
	Password   []byte
}

type AuthHTTPInitialPasswordInput struct {
	TenantSlug  string
	NewPassword []byte
}

type AuthHTTPInitialChallenge interface {
	ExpiresAt() time.Time
	RevealOnce(func(cookie, csrf []byte) error) error
	Destroy()
}

type AuthHTTPGrant interface {
	AccessToken() (string, error)
	AbsoluteExpiresAt() time.Time
	RevealRefreshOnce(func([]byte) error) error
	Destroy()
}

type AuthHTTPReauthenticationMethod string

const (
	AuthHTTPReauthenticationPassword AuthHTTPReauthenticationMethod = "password_v1"
	AuthHTTPReauthenticationWebAuthn AuthHTTPReauthenticationMethod = "webauthn_v1"
	AuthHTTPReauthenticationTOTP     AuthHTTPReauthenticationMethod = "totp_v1"
	AuthHTTPReauthenticationPolicy   int64                          = 1
)

type AuthHTTPReauthenticationStartInput struct {
	Method          AuthHTTPReauthenticationMethod
	AuthenticatorID uuid.UUID
	PurposeSHA256   [32]byte
	Origin          string
}

type AuthHTTPReauthenticationCompleteInput struct {
	ChallengeID                       uuid.UUID
	Challenge                         []byte
	Method                            AuthHTTPReauthenticationMethod
	PurposeSHA256                     [32]byte
	Origin                            string
	Password, WebAuthnAssertion, TOTP []byte
}

type AuthHTTPReauthenticationChallenge interface {
	ID() uuid.UUID
	Method() AuthHTTPReauthenticationMethod
	ExpiresAt() time.Time
	CredentialID() []byte
	RPID() string
	UserVerificationRequired() bool
	RevealOnce(func([]byte) error) error
	Destroy()
}

type AuthHTTPSession struct {
	UserID, TenantID, SessionID        uuid.UUID
	TenantSlug                         string
	Email, DisplayName                 string
	ClientID                           accesstoken.ClientID
	AccessMode                         accesstoken.AccessMode
	PortalBindingID                    uuid.UUID
	TokenVersion, AuthorizationVersion int64
	PlatformAuthorizationClock         int64
	PasswordCredentialRevision         *int64
	SecurityEpoch                      [32]byte
	DeploymentMode                     iamapplication.DeploymentMode
	TenantPurpose                      iamapplication.TenantPurpose
	PlatformCapabilities               []string
	TenantCapabilities                 []string
}

type AuthHTTPApplication interface {
	StartInitialPassword(
		context.Context, AuthHTTPLoginInput,
	) (AuthHTTPInitialChallenge, error)
	CompleteInitialPassword(
		context.Context, AuthHTTPInitialPasswordInput,
		*authentication.PresentedInitialPassword,
	) (AuthHTTPGrant, error)
	Login(context.Context, AuthHTTPLoginInput) (AuthHTTPGrant, error)
	Refresh(context.Context, *refreshtoken.Presented) (AuthHTTPGrant, error)
	Logout(context.Context, *refreshtoken.Presented) error
	ValidateRefresh(context.Context, *refreshtoken.Presented) error
	CurrentSession(context.Context, accesstoken.Claims) (AuthHTTPSession, error)
	StartReauthentication(
		context.Context, accesstoken.Claims, AuthHTTPReauthenticationStartInput,
	) (AuthHTTPReauthenticationChallenge, error)
	CompleteReauthentication(
		context.Context, accesstoken.Claims, AuthHTTPReauthenticationCompleteInput,
	) (AuthHTTPGrant, error)
}

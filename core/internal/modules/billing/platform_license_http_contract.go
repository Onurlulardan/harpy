package billing

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrPlatformLicenseHTTPAuthentication = errors.New("platform license authentication failed")
	ErrPlatformLicenseHTTPForbidden      = errors.New("platform license operation is forbidden")
	ErrPlatformLicenseHTTPUnavailable    = errors.New("platform license dependency is unavailable")
)

type PlatformLicenseHTTPEnvironment struct {
	EnvironmentID  string
	TenantID       string
	DeploymentMode string
}

func (environment PlatformLicenseHTTPEnvironment) valid() bool {
	environmentID, environmentErr := uuid.Parse(environment.EnvironmentID)
	tenantID, tenantErr := uuid.Parse(environment.TenantID)
	return environmentErr == nil && tenantErr == nil &&
		environmentID.Version() == uuid.Version(7) && tenantID != uuid.Nil &&
		(environment.DeploymentMode == "onpremise" ||
			environment.DeploymentMode == "saas")
}

type PlatformLicenseHTTPEnvironmentReader interface {
	Current(context.Context) (PlatformLicenseHTTPEnvironment, error)
}

type PlatformLicenseHTTPAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Capability          string
	ExpectedMode        string
}

type PlatformLicenseHTTPAccess struct {
	ActorID              string
	TenantID             string
	SessionID            string
	TokenVersion         int64
	AuthorizationVersion int64
	PlatformClock        int64
}

func (access PlatformLicenseHTTPAccess) valid() bool {
	actorID, actorErr := uuid.Parse(access.ActorID)
	tenantID, tenantErr := uuid.Parse(access.TenantID)
	sessionID, sessionErr := uuid.Parse(access.SessionID)
	return actorErr == nil && tenantErr == nil && sessionErr == nil &&
		actorID != uuid.Nil && tenantID != uuid.Nil && sessionID != uuid.Nil &&
		access.TokenVersion > 0 &&
		access.AuthorizationVersion > 0 && access.PlatformClock > 0
}

type PlatformLicenseHTTPAccessVerifier interface {
	Verify(
		context.Context,
		PlatformLicenseHTTPAccessRequest,
	) (PlatformLicenseHTTPAccess, error)
}

type PlatformLicenseHTTPDependencies struct {
	Access      PlatformLicenseHTTPAccessVerifier
	Environment PlatformLicenseHTTPEnvironmentReader
	Application PlatformLicenseApplication
	Report      ErrorSink
}

type PlatformLicenseHTTPHandler struct {
	access      PlatformLicenseHTTPAccessVerifier
	environment PlatformLicenseHTTPEnvironmentReader
	application PlatformLicenseApplication
	report      ErrorSink
}

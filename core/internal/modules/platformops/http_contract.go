package platformops

import (
	"context"
	"errors"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

var (
	ErrHTTPAuthentication = errors.New("platform operations HTTP authentication failed")
	ErrHTTPForbidden      = errors.New("platform operations HTTP operation is forbidden")
	ErrHTTPUnavailable    = errors.New("platform operations HTTP dependency is unavailable")
)

type HTTPEnvironment struct {
	EnvironmentID, TenantID, DeploymentMode string
}

func (environment HTTPEnvironment) valid() bool {
	return canonicalHTTPUUID(environment.EnvironmentID) &&
		canonicalHTTPUUID(environment.TenantID) &&
		(environment.DeploymentMode == "onpremise" || environment.DeploymentMode == "saas")
}

type HTTPEnvironmentReader interface {
	Current(context.Context) (HTTPEnvironment, error)
}

type HTTPAccessRequest struct {
	AuthorizationValues, TenantIDValues, TenantSlugValues []string
	Capability                                            string
	ExpectedMode                                          string
	RequireStepUp                                         bool
	StepUpPurpose                                         [32]byte
}

type HTTPAccess struct {
	ActorID, TenantID, SessionID                      string
	TokenVersion, AuthorizationVersion, PlatformClock int64
}

func (access HTTPAccess) valid() bool {
	return canonicalHTTPUUID(access.ActorID) && canonicalHTTPUUID(access.TenantID) &&
		canonicalHTTPUUID(access.SessionID) && access.TokenVersion > 0 &&
		access.AuthorizationVersion > 0 && access.PlatformClock > 0
}

type HTTPAccessVerifier interface {
	Verify(context.Context, HTTPAccessRequest) (HTTPAccess, error)
}

type HTTPDependencies struct {
	Access          HTTPAccessVerifier
	Environment     HTTPEnvironmentReader
	Diagnostics     platformopsapi.Diagnostics
	Protection      platformopsapi.RecoveryProtectionService
	Anchors         platformopsapi.RecoveryAnchors
	RecoverySets    platformopsapi.RecoverySets
	Restores        platformopsapi.Restores
	Operations      platformopsapi.HealthReader
	Traces          platformopsapi.TraceOperations
	Alerts          platformopsapi.AlertOperations
	PermissionCache platformopsapi.PermissionCacheOperations
	Report          func(error)
}

type HTTPHandler struct {
	access          HTTPAccessVerifier
	environment     HTTPEnvironmentReader
	diagnostics     platformopsapi.Diagnostics
	protection      platformopsapi.RecoveryProtectionService
	anchors         platformopsapi.RecoveryAnchors
	recoverySets    platformopsapi.RecoverySets
	restores        platformopsapi.Restores
	operations      platformopsapi.HealthReader
	traces          platformopsapi.TraceOperations
	alerts          platformopsapi.AlertOperations
	permissionCache platformopsapi.PermissionCacheOperations
	report          func(error)
}

func NewHTTPHandler(dependencies HTTPDependencies) (*HTTPHandler, error) {
	if nilHTTPPort(dependencies.Access) || nilHTTPPort(dependencies.Environment) ||
		nilHTTPPort(dependencies.Diagnostics) || nilHTTPPort(dependencies.Protection) ||
		nilHTTPPort(dependencies.Anchors) ||
		nilHTTPPort(dependencies.RecoverySets) || nilHTTPPort(dependencies.Restores) ||
		nilHTTPPort(dependencies.Operations) || nilHTTPPort(dependencies.Traces) ||
		nilHTTPPort(dependencies.Alerts) ||
		nilHTTPPort(dependencies.PermissionCache) ||
		dependencies.Report == nil {
		return nil, ErrHTTPUnavailable
	}
	return &HTTPHandler{access: dependencies.Access, environment: dependencies.Environment,
		diagnostics: dependencies.Diagnostics, protection: dependencies.Protection,
		anchors:      dependencies.Anchors,
		recoverySets: dependencies.RecoverySets, restores: dependencies.Restores,
		operations: dependencies.Operations, traces: dependencies.Traces,
		alerts:          dependencies.Alerts,
		permissionCache: dependencies.PermissionCache,
		report:          dependencies.Report}, nil
}

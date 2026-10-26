package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var currentSessionDomains = []string{"iam", "system_resource_layout"}

func (application *authHTTPApplication) CurrentSession(
	ctx context.Context, claims accesstoken.Claims,
) (AuthHTTPSession, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		claims.Subject() == uuid.Nil || claims.Tenant() == uuid.Nil ||
		claims.Session() == uuid.Nil {
		return AuthHTTPSession{}, ErrAuthHTTPInvalidSession
	}
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), currentSessionDomains,
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return AuthHTTPSession{}, ErrAuthHTTPUnavailable
	}
	request, err := buildCurrentSessionScope(
		claims.Subject(), claims.Tenant(), generations,
	)
	if err != nil {
		return AuthHTTPSession{}, ErrAuthHTTPUnavailable
	}
	var view sessionview.View
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var readErr error
			view, readErr = application.sessions.Read(
				scoped, scope, claims.Subject(), claims.Session(),
			)
			if readErr != nil {
				return readErr
			}
			view.PlatformCapabilities, readErr = application.platformCapabilities(
				scoped, scope, claims, view,
			)
			return readErr
		},
	)
	if err != nil {
		if errors.Is(err, sessionview.ErrRejected) {
			return AuthHTTPSession{}, ErrAuthHTTPInvalidSession
		}
		return AuthHTTPSession{}, ErrAuthHTTPUnavailable
	}
	result := sessionView(view)
	if !validSessionView(result, claims) {
		return AuthHTTPSession{}, ErrAuthHTTPInvalidSession
	}
	return result, nil
}

func buildCurrentSessionScope(
	userID uuid.UUID,
	tenantID uuid.UUID,
	generations map[string]int64,
) (platformtx.ScopeRequest, error) {
	if !validIAMLayoutSnapshotGenerations(generations) {
		return platformtx.ScopeRequest{}, errIAMHTTPSnapshotScope
	}
	actor, err := actorcontext.Issue(userID, tenantID)
	if err != nil {
		return platformtx.ScopeRequest{}, err
	}
	operation, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		"iam.sessionview.read", generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, err
	}
	return platformtx.ScopeRequest{
		Actor: actor, Operation: operation, Mode: platformtx.SnapshotQuery,
	}, nil
}

func sessionView(view sessionview.View) AuthHTTPSession {
	return AuthHTTPSession{
		UserID: view.UserID, TenantID: view.TenantID, SessionID: view.SessionID,
		TenantSlug: view.TenantSlug, Email: view.Email, DisplayName: view.DisplayName,
		ClientID:                   accesstoken.ClientID(view.ClientID),
		AccessMode:                 accesstoken.AccessMode(view.AccessMode),
		PortalBindingID:            view.PortalBindingID,
		TokenVersion:               view.TokenVersion,
		AuthorizationVersion:       view.AuthorizationVersion,
		PlatformAuthorizationClock: view.PlatformAuthorizationClock,
		PasswordCredentialRevision: view.PasswordCredentialRevision,
		SecurityEpoch:              view.SecurityEpoch,
		DeploymentMode:             view.DeploymentMode,
		TenantPurpose:              view.TenantPurpose,
		PlatformCapabilities:       append([]string{}, view.PlatformCapabilities...),
		TenantCapabilities:         append([]string{}, view.TenantCapabilities...),
	}
}

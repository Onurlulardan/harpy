package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

type authHTTPCurrentSessionResponse struct {
	SessionID                  uuid.UUID  `json:"session_id"`
	UserID                     uuid.UUID  `json:"user_id"`
	TenantID                   uuid.UUID  `json:"tenant_id"`
	TenantSlug                 string     `json:"tenant_slug"`
	Email                      string     `json:"email"`
	DisplayName                string     `json:"display_name"`
	ClientID                   string     `json:"client_id"`
	AccessMode                 string     `json:"access_mode"`
	PortalBindingID            *uuid.UUID `json:"portal_binding_id,omitempty"`
	TokenVersion               int64      `json:"token_version"`
	AuthorizationVersion       int64      `json:"authorization_version"`
	PasswordCredentialRevision *int64     `json:"password_credential_revision"`
	DeploymentMode             string     `json:"deployment_mode"`
	PlatformCapabilities       []string   `json:"platform_capabilities"`
	TenantCapabilities         []string   `json:"tenant_capabilities"`
	AuthenticationTime         time.Time  `json:"auth_time"`
	ExpiresAt                  time.Time  `json:"expires_at"`
	Assurance                  string     `json:"acr"`
	Methods                    []string   `json:"amr"`
}

func currentSessionResponse(
	view AuthHTTPSession,
	claims accesstoken.Claims,
) (authHTTPCurrentSessionResponse, bool) {
	if !validSessionView(view, claims) {
		return authHTTPCurrentSessionResponse{}, false
	}
	response := authHTTPCurrentSessionResponse{
		SessionID: view.SessionID, UserID: view.UserID,
		TenantID: view.TenantID, TenantSlug: view.TenantSlug,
		Email: view.Email, DisplayName: view.DisplayName,
		ClientID: string(view.ClientID), AccessMode: string(view.AccessMode),
		TokenVersion:               view.TokenVersion,
		AuthorizationVersion:       view.AuthorizationVersion,
		PasswordCredentialRevision: view.PasswordCredentialRevision,
		DeploymentMode:             string(view.DeploymentMode),
		PlatformCapabilities:       append([]string{}, view.PlatformCapabilities...),
		TenantCapabilities:         append([]string{}, view.TenantCapabilities...),
		AuthenticationTime:         time.Unix(claims.AuthenticationTimeUnix(), 0).UTC(),
		ExpiresAt:                  time.Unix(claims.ExpiresAtUnix(), 0).UTC(),
		Assurance:                  string(claims.Assurance()),
	}
	for _, method := range claims.Methods() {
		response.Methods = append(response.Methods, string(method))
	}
	if view.PortalBindingID != uuid.Nil {
		portalID := view.PortalBindingID
		response.PortalBindingID = &portalID
	}
	return response, true
}

package iam

import (
	"sort"
	"strings"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func validSessionView(view AuthHTTPSession, claims accesstoken.Claims) bool {
	if view.UserID == uuid.Nil || view.UserID != claims.Subject() ||
		view.TenantID == uuid.Nil || view.TenantID != claims.Tenant() ||
		view.SessionID == uuid.Nil || view.SessionID != claims.Session() ||
		!validLoginText(view.TenantSlug, 128) ||
		strings.TrimSpace(view.Email) == "" ||
		strings.TrimSpace(view.DisplayName) == "" ||
		view.ClientID != claims.Client() || view.AccessMode != claims.AccessMode() ||
		view.TokenVersion != claims.TokenVersion() ||
		view.AuthorizationVersion != claims.AuthorizationVersion() ||
		view.PlatformAuthorizationClock != claims.PlatformAuthorizationClock() ||
		view.SecurityEpoch != claims.SecurityEpoch() ||
		(view.PasswordCredentialRevision != nil &&
			*view.PasswordCredentialRevision < 1) ||
		!validPermissionSet(view.PlatformCapabilities) ||
		!validPermissionSet(view.TenantCapabilities) ||
		!(iamapplication.SessionCapabilitySets{
			Mode: view.DeploymentMode, Purpose: view.TenantPurpose,
			PlatformCapabilities: view.PlatformCapabilities,
			TenantCapabilities:   view.TenantCapabilities,
		}).Valid() {
		return false
	}
	portalID, hasPortal := claims.PortalBinding()
	if hasPortal {
		return portalID != uuid.Nil && view.PortalBindingID == portalID
	}
	return view.PortalBindingID == uuid.Nil
}

func validPermissionSet(values []string) bool {
	if values == nil || len(values) > 4096 || !sort.StringsAreSorted(values) {
		return false
	}
	for index, value := range values {
		if !validPermission(value) || index > 0 && values[index-1] == value {
			return false
		}
	}
	return true
}

func validPermission(value string) bool {
	if value == "" || len(value) > 128 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			character != '.' && character != '_' && character != ':' && character != '-' {
			return false
		}
	}
	return true
}

package platformauthority

import (
	"context"

	gcpauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/gcpkms"
	authorityapi "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

type InitialAuthorityPlanner = authorityapi.InitialAuthorityPlanner
type InitialAuthorityAcknowledger = authorityapi.InitialAuthorityAcknowledger
type InitialAuthorityRequest = authorityapi.InitialAuthorityRequest
type PreparedAuthority = authorityapi.PreparedAuthority
type AuthorityState = authorityapi.State
type AuthorityDeploymentMode = authorityapi.DeploymentMode
type AuthorityScopeKind = authorityapi.ScopeKind
type AuthorityExpiryPolicy = authorityapi.ExpiryPolicy
type GoogleClient = gcpauthority.GoogleClient

const (
	AuthorityStateSchema      = "protected-authority-state.v1"
	AuthorityModeSaaS         = authorityapi.ModeSaaS
	AuthorityScopeEnvironment = authorityapi.ScopeEnvironment
	AuthorityScopeTenant      = authorityapi.ScopeTenant
	AuthorityExpiryOnline     = authorityapi.ExpiryOnline
)

func OpenGoogleClient(
	ctx context.Context,
	workloadIdentityPath string,
) (*GoogleClient, error) {
	return gcpauthority.OpenGoogleClient(ctx, workloadIdentityPath)
}

package iam

import (
	"fmt"
	"net/url"

	recoveryanchorpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/recoveryanchor"
	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	recoveryanchorapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/recoveryanchor"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RecoveryAnchorRuntime struct{ inner *recoveryanchorapp.Runtime }

func NewRecoveryAnchorRuntime(
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	authority anchorapi.AuthorityPlanner,
	profile config.IAMHTTPProfile,
) (*RecoveryAnchorRuntime, error) {
	if pool == nil || coordinator == nil || authority == nil {
		return nil, fmt.Errorf("IAM recovery anchor dependencies are required")
	}
	store, err := recoveryanchorpostgres.NewStore()
	if err != nil {
		return nil, err
	}
	authorityAdapter, err := recoveryanchorapp.NewAuthorityPlanner(authority)
	if err != nil {
		return nil, err
	}
	service, err := recoveryanchorapp.NewService(store, authorityAdapter)
	if err != nil {
		return nil, err
	}
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	origins := profile.AllowedOrigins()
	parsed, err := url.Parse(origins[0])
	if err != nil || parsed.Hostname() == "" {
		return nil, fmt.Errorf("IAM recovery anchor RP is invalid")
	}
	runtime, err := recoveryanchorapp.NewRuntime(
		pool, coordinator, service, recoveryAnchorVerifier{}, parsed.Hostname(), origins)
	if err != nil {
		return nil, err
	}
	return &RecoveryAnchorRuntime{inner: runtime}, nil
}

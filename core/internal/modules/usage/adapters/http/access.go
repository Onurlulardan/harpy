package usagehttp

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

type AccessRequest struct {
	AuthorizationValues, TenantIDValues, TenantSlugValues []string
	Scope                                                 domain.Scope
	Permission, Operation                                 string
}

package v1bridge

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/google/uuid"
)

type tenantIdentityScope interface{ TenantID() uuid.UUID }

func tenantIdentity(scope api.Scope) (uuid.UUID, bool) {
	identity, ok := scope.(tenantIdentityScope)
	if !ok || identity.TenantID() == uuid.Nil {
		return uuid.Nil, false
	}
	return identity.TenantID(), true
}

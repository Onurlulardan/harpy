package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

type identityScope interface {
	TenantID() uuid.UUID
	ActorID() uuid.UUID
}

func domainID(value [16]byte) (domain.ID, error) {
	return domain.IDFromBytes(value[:])
}

func wireID(value domain.ID) api.ID {
	return api.ID(value)
}

func scopeIDs(scope api.Scope) (domain.ID, domain.ID, error) {
	identity, ok := scope.(identityScope)
	if !ok || identity.TenantID() == uuid.Nil || identity.ActorID() == uuid.Nil {
		return domain.ID{}, domain.ID{}, ErrInvalid
	}
	tenant, tenantErr := domainID(identity.TenantID())
	actor, actorErr := domainID(identity.ActorID())
	if tenantErr != nil || actorErr != nil {
		return domain.ID{}, domain.ID{}, ErrInvalid
	}
	return tenant, actor, nil
}

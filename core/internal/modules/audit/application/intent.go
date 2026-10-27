package application

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

// AppendIntent is an internal owner-to-audit contract. HTTP transports must not expose it.
type AppendIntent struct {
	OwnerAuditID         uuid.UUID
	OrganizationID       *uuid.UUID
	CorrelationID        *uuid.UUID
	OwnerModule          string
	OwnerOperation       string
	EntityType           string
	EntityID             string
	FormSlug             string
	Action               domain.Action
	AuthorizationVersion int64
	ReasonCode           string
	FieldChanges         []domain.FieldChange
	OccurredAt           time.Time
}

func domainID(value uuid.UUID) domain.ID { return domain.ID(value) }

func optionalDomainID(value *uuid.UUID) *domain.ID {
	if value == nil {
		return nil
	}
	converted := domainID(*value)
	return &converted
}

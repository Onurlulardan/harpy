package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	"github.com/google/uuid"
)

type scanner interface{ Scan(...any) error }

func scanEntry(source scanner, notFound func(error) bool, trailing ...any) (domain.Entry, error) {
	var entry domain.Entry
	var id, ownerAuditID, actorID uuid.UUID
	var organizationID, correlationID *uuid.UUID
	var formSlug, reasonCode *string
	var fieldChanges, eventSHA256 []byte
	arguments := []any{&id, &ownerAuditID, &actorID,
		&organizationID, &correlationID, &entry.OwnerModule, &entry.Operation,
		&entry.EntityType, &entry.EntityID, &formSlug, &entry.Action,
		&entry.AuthorizationVersion, &reasonCode, &fieldChanges, &eventSHA256,
		&entry.OccurredAt, &entry.RecordedAt}
	arguments = append(arguments, trailing...)
	if err := source.Scan(arguments...); err != nil {
		return domain.Entry{}, mapError(err, notFound)
	}
	entry.ID, entry.OwnerAuditID, entry.ActorID = domain.ID(id), domain.ID(ownerAuditID), domain.ID(actorID)
	entry.OrganizationID = domainIDPointer(organizationID)
	entry.CorrelationID = domainIDPointer(correlationID)
	if formSlug != nil {
		entry.FormSlug = *formSlug
	}
	if reasonCode != nil {
		entry.ReasonCode = *reasonCode
	}
	if len(eventSHA256) != 32 {
		return domain.Entry{}, application.ErrUnavailable
	}
	copy(entry.EventSHA256[:], eventSHA256)
	changes, err := decodeFieldChanges(fieldChanges)
	if err != nil {
		return domain.Entry{}, err
	}
	entry.FieldChanges = changes
	entry.OccurredAt = entry.OccurredAt.UTC().Truncate(time.Microsecond)
	entry.RecordedAt = entry.RecordedAt.UTC().Truncate(time.Microsecond)
	if entry.Validate() != nil {
		return domain.Entry{}, application.ErrUnavailable
	}
	return entry, nil
}

func uuidArgument(id domain.ID) uuid.UUID { return uuid.UUID(id) }

func optionalUUIDArgument(id *domain.ID) any {
	if id == nil {
		return nil
	}
	return uuidArgument(*id)
}

func domainIDPointer(id *uuid.UUID) *domain.ID {
	if id == nil {
		return nil
	}
	converted := domain.ID(*id)
	return &converted
}

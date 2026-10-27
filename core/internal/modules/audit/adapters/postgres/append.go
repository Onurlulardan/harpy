package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

func (store *Store) Append(
	ctx context.Context, mutation application.AppendMutation,
) (domain.Entry, bool, error) {
	if !store.allowed(ctx, application.ScopeCommand) || mutation.Entry.Validate() != nil ||
		mutation.RequestSHA256 == [32]byte{} ||
		mutation.RequestSHA256 != mutation.Entry.EventSHA256 {
		return domain.Entry{}, false, application.ErrInvalid
	}
	fieldChanges, err := encodeFieldChanges(mutation.Entry.FieldChanges)
	if err != nil {
		return domain.Entry{}, false, err
	}
	entry := mutation.Entry
	var replay bool
	persisted, err := scanEntry(store.queryRow(ctx, appendSQL,
		uuidArgument(entry.ID), uuidArgument(entry.OwnerAuditID), uuidArgument(entry.ActorID),
		optionalUUIDArgument(entry.OrganizationID), optionalUUIDArgument(entry.CorrelationID),
		entry.OwnerModule, entry.Operation, entry.EntityType,
		entry.EntityID, optionalText(entry.FormSlug), string(entry.Action),
		entry.AuthorizationVersion, optionalText(entry.ReasonCode), fieldChanges,
		entry.EventSHA256[:], entry.OccurredAt, entry.RecordedAt), store.notFound, &replay)
	if err != nil {
		return domain.Entry{}, false, err
	}
	if persisted.OwnerAuditID != entry.OwnerAuditID ||
		persisted.EventSHA256 != mutation.RequestSHA256 {
		return domain.Entry{}, false, application.ErrUnavailable
	}
	return persisted, replay, nil
}

func optionalText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (store *Store) allowed(ctx context.Context, access application.ScopeAccess) bool {
	return store != nil && store.usable != nil && store.queryRow != nil &&
		store.query != nil && store.scopeAccess == access && store.usable(ctx)
}

var _ application.Store = (*Store)(nil)

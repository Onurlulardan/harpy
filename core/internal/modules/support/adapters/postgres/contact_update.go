package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) ContactUpdate(ctx context.Context, access application.Access,
	update domain.ContactUpdate,
) (domain.ContactMessage, bool, error) {
	if !access.Allows(application.ScopePlatform, application.PermissionPlatformContactManage,
		application.OperationContactUpdate) || update.ActorID != domain.ID(access.ActorID()) {
		return domain.ContactMessage{}, false, application.ErrForbidden
	}
	var message domain.ContactMessage
	var replayed bool
	err := store.within(ctx, CommandTransaction, func(database Database) error {
		var document []byte
		err := database.QueryRow(ctx, updateContactSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), uuid.UUID(update.ID),
			update.Status != nil, contactStatusValue(update.Status), update.Notes != nil,
			stringValue(update.Notes), update.ExpectedRevision, update.IdempotencyKey,
			update.RequestSHA256[:], update.ObservedAt).Scan(&document, &replayed)
		if err != nil {
			return err
		}
		message, err = decodeContact(document)
		return err
	})
	return message, replayed, store.normalize(err)
}

func contactStatusValue(value *domain.ContactStatus) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

func stringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

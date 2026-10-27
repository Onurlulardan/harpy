package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) ContactGet(ctx context.Context, access application.Access,
	id domain.ID,
) (domain.ContactMessage, error) {
	if !access.Allows(application.ScopePlatform, application.PermissionPlatformContactRead,
		application.OperationContactRead) {
		return domain.ContactMessage{}, application.ErrForbidden
	}
	var message domain.ContactMessage
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, getContactSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), uuid.UUID(id)).Scan(&document); err != nil {
			return err
		}
		var err error
		message, err = decodeContact(document)
		return err
	})
	return message, store.normalize(err)
}

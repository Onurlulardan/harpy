package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) ContactsList(ctx context.Context, access application.Access,
	filter domain.ContactFilter,
) (domain.ContactPage, error) {
	if !access.Allows(application.ScopePlatform, application.PermissionPlatformContactRead,
		application.OperationContactList) {
		return domain.ContactPage{}, application.ErrForbidden
	}
	var page domain.ContactPage
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, listContactsSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), nullableText(string(filter.Status)),
			nullableText(filter.ContactType), filter.Page.Number, filter.Page.Size).
			Scan(&document, &page.Total); err != nil {
			return err
		}
		items, err := decodeContactPage(document)
		page.Items, page.Page = items, filter.Page
		return err
	})
	return page, store.normalize(err)
}

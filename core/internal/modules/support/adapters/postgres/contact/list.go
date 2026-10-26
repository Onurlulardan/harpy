package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) ListContacts(ctx context.Context, access ports.Access,
	filter domain.ContactFilter,
) (domain.ContactPage, error) {
	if !store.ready() {
		return domain.ContactPage{}, ports.ErrUnavailable
	}
	return store.driver.ContactsList(ctx, access, filter)
}

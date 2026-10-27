package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) UpdateContact(ctx context.Context, access ports.Access,
	update domain.ContactUpdate,
) (domain.ContactMessage, bool, error) {
	if !store.ready() {
		return domain.ContactMessage{}, false, ports.ErrUnavailable
	}
	return store.driver.ContactUpdate(ctx, access, update)
}

package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) GetContact(ctx context.Context, access ports.Access,
	id domain.ID,
) (domain.ContactMessage, error) {
	if !store.ready() {
		return domain.ContactMessage{}, ports.ErrUnavailable
	}
	return store.driver.ContactGet(ctx, access, id)
}

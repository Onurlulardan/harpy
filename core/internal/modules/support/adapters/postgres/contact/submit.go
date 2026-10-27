package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (store *Store) SubmitContact(ctx context.Context,
	command ports.ContactSubmission,
) (domain.ContactMessage, bool, error) {
	if !store.ready() {
		return domain.ContactMessage{}, false, ports.ErrUnavailable
	}
	return store.driver.ContactSubmit(ctx, command)
}

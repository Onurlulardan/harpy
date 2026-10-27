package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) ListContacts(ctx context.Context, access ports.Access,
	filter domain.ContactFilter,
) (domain.ContactPage, error) {
	if !service.ready() {
		return domain.ContactPage{}, ports.ErrUnavailable
	}
	return service.kernel.ContactsList(ctx, access, filter)
}

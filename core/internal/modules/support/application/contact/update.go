package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) UpdateContact(ctx context.Context, access ports.Access,
	update domain.ContactUpdate,
) (domain.ContactMessage, error) {
	if !service.ready() {
		return domain.ContactMessage{}, ports.ErrUnavailable
	}
	return service.kernel.ContactUpdate(ctx, access, update)
}

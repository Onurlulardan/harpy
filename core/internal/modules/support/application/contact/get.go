package contact

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) GetContact(ctx context.Context, access ports.Access,
	id domain.ID,
) (domain.ContactMessage, error) {
	if !service.ready() {
		return domain.ContactMessage{}, ports.ErrUnavailable
	}
	return service.kernel.ContactGet(ctx, access, id)
}

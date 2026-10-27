package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ContactGet(ctx context.Context, access Access,
	id domain.ID,
) (domain.ContactMessage, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformContactRead, OperationContactRead) {
		return domain.ContactMessage{}, ErrForbidden
	}
	if id == domain.NilID {
		return domain.ContactMessage{}, ErrInvalid
	}
	message, err := service.store.ContactGet(ctx, access, id)
	return message, normalizeStoreError(err)
}

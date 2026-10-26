package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ContactsList(ctx context.Context, access Access,
	filter domain.ContactFilter,
) (domain.ContactPage, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformContactRead, OperationContactList) {
		return domain.ContactPage{}, ErrForbidden
	}
	if !validContactFilter(filter) {
		return domain.ContactPage{}, ErrInvalid
	}
	page, err := service.store.ContactsList(ctx, access, filter)
	return page, normalizeStoreError(err)
}

func validContactFilter(filter domain.ContactFilter) bool {
	return filter.Page.Valid() && (filter.Status == "" || domain.ValidContactStatus(filter.Status)) &&
		(filter.ContactType == "" || domainText(filter.ContactType, 1, 50))
}

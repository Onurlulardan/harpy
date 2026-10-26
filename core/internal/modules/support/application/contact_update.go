package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ContactUpdate(ctx context.Context, access Access,
	update domain.ContactUpdate,
) (domain.ContactMessage, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformContactManage, OperationContactUpdate) ||
		update.ActorID != domain.ID(access.ActorID()) {
		return domain.ContactMessage{}, ErrForbidden
	}
	if !validContactUpdate(update) {
		return domain.ContactMessage{}, ErrInvalid
	}
	now, err := service.now()
	if err != nil {
		return domain.ContactMessage{}, err
	}
	update.ObservedAt = now
	message, _, err := service.store.ContactUpdate(ctx, access, update)
	return message, normalizeStoreError(err)
}

func validContactUpdate(update domain.ContactUpdate) bool {
	if update.ID == domain.NilID || update.ActorID == domain.NilID || update.ExpectedRevision < 1 ||
		!validCommandEvidence(update.IdempotencyKey, update.RequestSHA256) ||
		(update.Status == nil && update.Notes == nil) {
		return false
	}
	if update.Status != nil && !domain.ValidContactStatus(*update.Status) {
		return false
	}
	return update.Notes == nil || domainText(strings.TrimSpace(*update.Notes), 0, 8000)
}

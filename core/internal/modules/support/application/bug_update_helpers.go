package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) updateBug(ctx context.Context, access Access,
	update domain.BugUpdate,
) (domain.BugReport, error) {
	update.ActorID = domain.ID(access.ActorID())
	if !validBugUpdate(update) {
		return domain.BugReport{}, ErrInvalid
	}
	now, err := service.now()
	if err != nil {
		return domain.BugReport{}, err
	}
	update.ObservedAt = now
	report, _, err := service.store.BugUpdate(ctx, access, update)
	return report, normalizeStoreError(err)
}

func validBugUpdate(update domain.BugUpdate) bool {
	if update.ID == domain.NilID || update.TenantID == domain.NilID || update.ActorID == domain.NilID ||
		update.ExpectedRevision < 1 || !validCommandEvidence(update.IdempotencyKey, update.RequestSHA256) ||
		(update.Status == nil && update.Priority == nil && update.Resolution == nil && update.AdminNotes == nil) {
		return false
	}
	if update.Status != nil && !domain.ValidBugStatus(*update.Status) {
		return false
	}
	if update.Priority != nil && !domain.ValidBugPriority(*update.Priority) {
		return false
	}
	return (update.Resolution == nil || domainText(strings.TrimSpace(*update.Resolution), 0, 8000)) &&
		(update.AdminNotes == nil || domainText(strings.TrimSpace(*update.AdminNotes), 0, 8000))
}

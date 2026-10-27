package application

import (
	"context"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) BugCreate(ctx context.Context, access Access,
	input domain.BugInput, idempotencyKey string, requestSHA256 [32]byte,
) (domain.BugReport, error) {
	if !access.Allows(ScopeReporter, PermissionBugCreate, OperationBugCreate) {
		return domain.BugReport{}, ErrForbidden
	}
	if !validCommandEvidence(idempotencyKey, requestSHA256) {
		return domain.BugReport{}, ErrInvalid
	}
	normalized, err := domain.NormalizeBug(input, canonicaljsonv1.Canonicalize)
	if err != nil {
		return domain.BugReport{}, ErrInvalid
	}
	id, err := service.newID()
	if err != nil {
		return domain.BugReport{}, err
	}
	now, err := service.now()
	if err != nil {
		return domain.BugReport{}, err
	}
	report, _, err := service.store.BugCreate(ctx, BugSubmission{ID: id, Access: access,
		Input: normalized, IdempotencyKey: idempotencyKey,
		RequestSHA256: requestSHA256, ObservedAt: now})
	return reporterView(report), normalizeStoreError(err)
}

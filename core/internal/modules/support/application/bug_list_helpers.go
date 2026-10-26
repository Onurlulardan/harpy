package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) listBugs(ctx context.Context, access Access,
	filter domain.BugFilter, redact bool,
) (domain.BugPage, error) {
	if !validBugFilter(filter) {
		return domain.BugPage{}, ErrInvalid
	}
	page, err := service.store.BugsList(ctx, access, filter)
	if redact {
		for index := range page.Items {
			page.Items[index] = reporterView(page.Items[index])
		}
	}
	return page, normalizeStoreError(err)
}

func validBugFilter(filter domain.BugFilter) bool {
	return filter.Page.Valid() && (filter.Status == "" || domain.ValidBugStatus(filter.Status)) &&
		(filter.Severity == "" || domain.ValidBugSeverity(filter.Severity)) &&
		(filter.Category == "" || domain.ValidBugCategory(filter.Category))
}

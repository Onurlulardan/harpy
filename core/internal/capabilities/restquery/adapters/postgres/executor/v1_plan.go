package executor

import (
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func validateV1PagePlan(
	plan resolved.AuthorizedPlan, pagePlan v1query.PagePlan,
) (v1query.PredecessorOrdinal, bool, error) {
	pageNumber, pageErr := pagePlan.Page()
	limit, limitErr := pagePlan.Limit()
	ordinal, hasPredecessor, ordinalErr := pagePlan.PredecessorOrdinal()
	if plan.Validate() != nil || pageErr != nil || limitErr != nil || ordinalErr != nil {
		return v1query.PredecessorOrdinal{}, false, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return v1query.PredecessorOrdinal{}, false, ErrContract
	}
	mode, err := spec.Mode()
	query, found, queryErr := spec.Record()
	if err != nil || mode != api.QueryModeRecord || queryErr != nil || !found {
		return v1query.PredecessorOrdinal{}, false, ErrContract
	}
	page, err := query.Page()
	if err != nil {
		return v1query.PredecessorOrdinal{}, false, ErrContract
	}
	size, sizeErr := page.Size()
	_, hasAfter, afterErr := page.After()
	count, countErr := query.Count()
	if sizeErr != nil || afterErr != nil || countErr != nil || hasAfter ||
		size != limit || count != api.CountModeExact ||
		hasPredecessor != (pageNumber > 1) {
		return v1query.PredecessorOrdinal{}, false, ErrContract
	}
	if !hasPredecessor {
		return v1query.PredecessorOrdinal{}, false, nil
	}
	return ordinal, true, nil
}

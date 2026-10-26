package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func resolveExactCountCapability(
	request api.QueryRequest,
	catalog api.FieldCatalog,
	collector *resolutionIssues,
) bool {
	capabilities, err := catalog.Capabilities()
	if err != nil {
		collector.countBudget("/count", restqueryv1.IssueCapabilityRequired)
		return false
	}
	enabled, err := capabilities.ExactCountEnabled()
	if err != nil || !enabled {
		if request.Count() == api.CountModeExact {
			collector.countBudget("/count", restqueryv1.IssueCapabilityRequired)
		}
		return false
	}
	if request.Count() != api.CountModeExact {
		return true
	}
	if where, found := request.Where(); found {
		if where.Depth() > 3 {
			collector.countBudget("/where", restqueryv1.IssueBooleanDepthExceeded)
		}
		if requestPredicateCount(where) > 20 {
			collector.countBudget("/where", restqueryv1.IssuePredicateCountExceeded)
		}
		checkCountSafeFilter(catalog, where, "/where", collector)
	}
	if _, found := request.Search(); found {
		checkCountSafeSearch(catalog, collector)
	}
	return true
}

func requestPredicateCount(filter api.RequestFilter) int {
	if _, found := filter.Predicate(); found {
		return 1
	}
	count := 0
	for _, child := range filter.Children() {
		count += requestPredicateCount(child)
	}
	return count
}

func checkCountSafeFilter(
	catalog api.FieldCatalog,
	filter api.RequestFilter,
	path string,
	collector *resolutionIssues,
) {
	if predicate, found := filter.Predicate(); found {
		field, fieldFound, err := catalog.FieldByPublicName(predicate.Field())
		if err != nil || !fieldFound {
			return
		}
		capabilities, err := field.Capabilities()
		if err != nil {
			return
		}
		capability, allowed, err := capabilities.Filter(predicate.Operator())
		if err != nil || !allowed {
			return
		}
		countSafe, err := capability.CountSafe()
		if err != nil || !countSafe {
			collector.countBudget(path+"/op", restqueryv1.IssueOperatorNotAllowed)
		}
		return
	}
	for index, child := range filter.Children() {
		checkCountSafeFilter(
			catalog, child, childFilterPath(path, filter.Kind(), index), collector,
		)
	}
}

func checkCountSafeSearch(catalog api.FieldCatalog, collector *resolutionIssues) {
	fieldIDs, err := catalog.SearchFields()
	if err != nil {
		return
	}
	for _, fieldID := range fieldIDs {
		field, found, fieldErr := catalog.Field(fieldID)
		if fieldErr != nil || !found {
			return
		}
		capabilities, capabilityErr := field.Capabilities()
		if capabilityErr != nil {
			return
		}
		capability, allowed, filterErr := capabilities.Filter(api.OperatorContains)
		if filterErr != nil || !allowed {
			collector.countBudget("/search", restqueryv1.IssueOperatorNotAllowed)
			return
		}
		countSafe, countErr := capability.CountSafe()
		if countErr != nil || !countSafe {
			collector.countBudget("/search", restqueryv1.IssueOperatorNotAllowed)
			return
		}
	}
}

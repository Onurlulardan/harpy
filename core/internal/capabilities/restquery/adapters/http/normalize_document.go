package httpquery

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/contract"
)

func normalizeQueryDocument(
	document contract.QueryDocument, collector *issueCollector,
) (api.QueryRequest, bool) {
	aggregate, present := memberValue(document.Aggregate())
	if present && !rawIsNull(aggregate) {
		return normalizeAggregateDocument(document, aggregate, collector)
	}
	return normalizeRecordDocument(document, collector)
}

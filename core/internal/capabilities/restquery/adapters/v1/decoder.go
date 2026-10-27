package v1query

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (decoder decoder) Decode(rawQuery string) (pendingResult, error) {
	return decoder.decode(rawQuery, DialectDynamicList)
}

func (decoder decoder) decode(
	rawQuery string, dialect Dialect,
) (pendingResult, error) {
	if err := decoder.Validate(); err != nil {
		return pendingResult{}, err
	}
	if err := dialect.Validate(); err != nil {
		return pendingResult{}, err
	}
	leaves, err := decodeWire(rawQuery)
	if err != nil {
		return pendingResult{}, err
	}
	collector := &issueCollector{}
	var normalized []normalizedLeaf
	if dialect == DialectOpenAPIGet {
		normalized = normalizeOpenAPILeaves(leaves, decoder.manifest, collector)
	} else {
		normalized = normalizeQueryLeaves(leaves, decoder.manifest, collector)
	}
	document := decodeDocument(normalized, collector)
	pagePlan := decodePagination(document, decoder.manifest, collector)
	where := decodeFilters(document, decoder.manifest, collector)
	search := decodeSearch(document, collector)
	sorts := decodeSort(document, decoder.manifest, collector)
	if collector.hasIssues() {
		return pendingResult{}, collector.queryError()
	}
	page, err := api.NewRequestPage(pagePlan.limit, nil)
	if err != nil {
		return pendingResult{}, fmt.Errorf("build V1 REST Query page: %w", err)
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{}, Where: where, Sort: sorts, Page: page,
		Search: search, Count: api.CountModeExact, Include: []api.RequestInclude{},
	})
	if err != nil {
		return pendingResult{}, fmt.Errorf("build V1 REST Query request: %w", err)
	}
	result, err := newPendingResult(
		request, pagePlan, decoder.manifest.id, document.aliases.list(),
	)
	if err != nil {
		return pendingResult{}, fmt.Errorf("build V1 REST Query result: %w", err)
	}
	return result, nil
}

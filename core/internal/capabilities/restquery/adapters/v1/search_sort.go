package v1query

import (
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"golang.org/x/text/unicode/norm"
)

func decodeSearch(
	document v1Document, collector *issueCollector,
) *api.RequestSearch {
	if document.search == nil || document.search.value == "" {
		return nil
	}
	value := document.search.value
	normalized := norm.NFC.String(value)
	if utf8.RuneCountInString(normalized) > 128 || len(normalized) > 256 {
		collector.addBudget(document.search.path, restqueryv1.IssueSearchSizeExceeded)
		return nil
	}
	search, err := api.NewRequestSearch(value)
	if err != nil {
		collector.add(document.search.path, restqueryv1.IssueValueNotAllowed)
		return nil
	}
	return &search
}

func decodeSort(
	document v1Document, policy manifest, collector *issueCollector,
) []api.RequestSort {
	sort := policy.defaultSort
	valid := true
	if document.sortField != nil && document.sortField.value != "" {
		binding, found := policy.sortFields[document.sortField.value]
		if !found {
			collector.add(document.sortField.path, restqueryv1.IssueFieldUnknown)
			valid = false
		} else {
			document.aliases.add(
				binding.aliasID, document.sortField.path,
				restqueryv1.IssueFieldUnknown,
			)
			updated, err := api.NewRequestSort(binding.name, sort.Direction())
			if err != nil {
				collector.add(document.sortField.path, restqueryv1.IssueValueNotAllowed)
				valid = false
			} else {
				sort = updated
			}
		}
	}
	if document.sortOrder != nil && document.sortOrder.value != "" {
		direction, found := api.ParseDirection(strings.ToLower(document.sortOrder.value))
		if !found {
			collector.add(document.sortOrder.path, restqueryv1.IssueValueNotAllowed)
			valid = false
		} else {
			updated, err := api.NewRequestSort(sort.Field(), direction)
			if err != nil {
				collector.add(document.sortOrder.path, restqueryv1.IssueValueNotAllowed)
				valid = false
			} else {
				sort = updated
			}
		}
	}
	if !valid {
		return nil
	}
	return []api.RequestSort{sort}
}

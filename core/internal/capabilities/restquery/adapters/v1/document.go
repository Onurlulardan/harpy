package v1query

import (
	"slices"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

type documentMember struct {
	value string
	path  string
}

type rawFilter struct {
	field    string
	path     string
	scalar   *documentMember
	operator *documentMember
	value    *documentMember
	array    []documentMember
}

type v1Document struct {
	page           *documentMember
	limit          *documentMember
	search         *documentMember
	filterOperator *documentMember
	sortField      *documentMember
	sortOrder      *documentMember
	filters        map[string]*rawFilter
	aliases        aliasSet
}

func decodeDocument(
	leaves []normalizedLeaf, collector *issueCollector,
) v1Document {
	document := v1Document{filters: make(map[string]*rawFilter), aliases: newAliasSet()}
	for _, leaf := range leaves {
		document.aliases.add(
			leaf.aliasID, leaf.path, restqueryv1.IssueMemberUnknown,
		)
		member := documentMember{value: leaf.value, path: leaf.path}
		if len(leaf.segments) == 1 {
			document.setSimple(leaf.segments[0], member, collector)
			continue
		}
		if leaf.segments[0] == "sort" {
			document.setSort(leaf.segments[1], member)
			continue
		}
		document.setFilter(leaf.segments, member, collector)
	}
	return document
}

func (document *v1Document) setSimple(
	key string, member documentMember, collector *issueCollector,
) {
	var target **documentMember
	switch key {
	case "page":
		target = &document.page
	case "limit":
		target = &document.limit
	case "search":
		target = &document.search
	case "filter_operator":
		target = &document.filterOperator
	}
	if target == nil {
		return
	}
	if *target != nil && (*target).path != member.path {
		collector.add(member.path, restqueryv1.IssueMemberDuplicate)
		return
	}
	setFirst(target, member)
}

func (document *v1Document) setSort(key string, member documentMember) {
	if key == "field" {
		setFirst(&document.sortField, member)
	} else {
		setFirst(&document.sortOrder, member)
	}
}

func (document *v1Document) setFilter(
	segments []string, member documentMember, collector *issueCollector,
) {
	field := segments[1]
	filter, found := document.filters[field]
	if !found {
		filter = &rawFilter{field: field, path: appendPointer("/filters", field)}
		document.filters[field] = filter
	}
	if len(segments) == 2 {
		setFirst(&filter.scalar, member)
	} else if segments[2] == "operator" {
		setFirst(&filter.operator, member)
	} else if segments[2] == "value" {
		setFirst(&filter.value, member)
	} else {
		filter.array = append(filter.array, member)
	}
	if filter.scalar != nil &&
		(filter.operator != nil || filter.value != nil || len(filter.array) != 0) ||
		len(filter.array) != 0 && (filter.operator != nil || filter.value != nil) {
		collector.add(filter.path, restqueryv1.IssueMemberDuplicate)
	}
}

func (document v1Document) orderedFilters() []*rawFilter {
	keys := make([]string, 0, len(document.filters))
	for key := range document.filters {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := make([]*rawFilter, 0, len(keys))
	for _, key := range keys {
		result = append(result, document.filters[key])
	}
	return result
}

func setFirst(target **documentMember, value documentMember) {
	if *target == nil {
		*target = &value
	}
}

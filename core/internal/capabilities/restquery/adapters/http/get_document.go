package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

type getMember struct {
	value   string
	pointer string
}

type getFilterDocument struct {
	field    *getMember
	operator *getMember
	value    *getMember
	values   map[int]getMember
}

type getSortDocument struct {
	field     *getMember
	direction *getMember
}

type getDocument struct {
	selects   map[int]getMember
	filters   map[int]*getFilterDocument
	sorts     map[int]*getSortDocument
	search    *getMember
	pageSize  *getMember
	pageAfter *getMember
	count     *getMember
}

func newGETDocument() getDocument {
	return getDocument{
		selects: make(map[int]getMember), filters: make(map[int]*getFilterDocument),
		sorts: make(map[int]*getSortDocument),
	}
}

func decodeGETDocument(leaves []getLeaf, collector *issueCollector) getDocument {
	document := newGETDocument()
	for _, leaf := range leaves {
		if leaf.segments == nil {
			collector.add("", restqueryv1.IssueValueNotCanonical)
			continue
		}
		if !leaf.structuralNFC {
			collector.add(leaf.pointer, restqueryv1.IssueNFCRequired)
			continue
		}
		switch leaf.segments[0] {
		case "select":
			document.addGETSelect(leaf, collector)
		case "filter":
			document.addGETFilter(leaf, collector)
		case "sort":
			document.addGETSort(leaf, collector)
		case "search", "page", "count":
			document.addGETScalar(leaf, collector)
		case "include", "aggregate":
			collector.add(
				appendJSONPointer("", leaf.segments[0]),
				restqueryv1.IssueMemberForbidden,
			)
		default:
			collector.add(leaf.pointer, restqueryv1.IssueMemberUnknown)
		}
	}
	return document
}

func memberFromGETLeaf(leaf getLeaf) *getMember {
	return &getMember{value: leaf.value, pointer: leaf.pointer}
}

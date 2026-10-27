package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

type publishedGETParameter struct {
	single *getMember
	list   map[int]getMember
}

type publishedGETDocument struct {
	parameters map[string]*publishedGETParameter
	pageSize   *getMember
	pageAfter  *getMember
}

func decodePublishedGETDocument(
	leaves []getLeaf, collector *publishedIssueCollector,
) publishedGETDocument {
	document := publishedGETDocument{
		parameters: make(map[string]*publishedGETParameter),
	}
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
		case "param":
			document.addPublishedGETParameter(leaf, collector)
		case "page":
			document.addPublishedGETPage(leaf, collector)
		default:
			collector.add(leaf.pointer, restqueryv1.IssueParameterUnknown)
		}
	}
	return document
}

func (document *publishedGETDocument) addPublishedGETParameter(
	leaf getLeaf, collector *publishedIssueCollector,
) {
	if len(leaf.segments) < 2 || leaf.segments[1] == "" {
		collector.add("/param", restqueryv1.IssueParameterUnknown)
		return
	}
	name := leaf.segments[1]
	parameter := document.parameters[name]
	if parameter == nil {
		parameter = &publishedGETParameter{}
		document.parameters[name] = parameter
	}
	if len(leaf.segments) == 2 {
		parameter.single = memberFromGETLeaf(leaf)
		return
	}
	if len(leaf.segments) != 3 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	index, valid := decodePublishedGETIndex(
		leaf.segments[2], leaf.pointer, collector,
	)
	if !valid {
		return
	}
	if parameter.list == nil {
		parameter.list = make(map[int]getMember)
	}
	parameter.list[index] = *memberFromGETLeaf(leaf)
}

func (document *publishedGETDocument) addPublishedGETPage(
	leaf getLeaf, collector *publishedIssueCollector,
) {
	if len(leaf.segments) != 2 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	member := memberFromGETLeaf(leaf)
	switch leaf.segments[1] {
	case "size":
		document.pageSize = member
	case "after":
		document.pageAfter = member
	default:
		collector.add(leaf.pointer, restqueryv1.IssueParameterUnknown)
	}
}

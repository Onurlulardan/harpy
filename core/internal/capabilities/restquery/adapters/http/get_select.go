package httpquery

import restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"

func (document *getDocument) addGETSelect(
	leaf getLeaf, collector *issueCollector,
) {
	if len(leaf.segments) < 2 {
		collector.add("/select", restqueryv1.IssueIndexInvalid)
		return
	}
	indexPath := appendJSONPointer("/select", leaf.segments[1])
	index, valid := decodeGETIndex(leaf.segments[1], indexPath, collector)
	if !valid {
		return
	}
	if len(leaf.segments) != 2 {
		collector.add(leaf.pointer, restqueryv1.IssueValueNotCanonical)
		return
	}
	document.selects[index] = *memberFromGETLeaf(leaf)
}

func decodeGETSelects(
	document getDocument, collector *issueCollector,
) ([]string, bool) {
	indices, valid := orderedGETIndices(document.selects, "/select", collector)
	if len(indices) > 50 {
		collector.addBudget("/select", restqueryv1.IssueSelectCountExceeded)
		valid = false
	}
	result := make([]string, 0, len(indices))
	seen := make(map[string]struct{}, len(indices))
	for _, index := range indices {
		member := document.selects[index]
		name, nameValid := decodeGETFieldName(member, collector)
		if !nameValid {
			valid = false
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			collector.add(member.pointer, restqueryv1.IssueFieldDuplicate)
			valid = false
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result, valid
}

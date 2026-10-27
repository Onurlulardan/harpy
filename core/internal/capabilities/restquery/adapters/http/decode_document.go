package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/contract"
)

var rootMembers = []string{
	"select", "where", "sort", "page", "search", "count", "include", "aggregate",
}

func decodeQueryDocument(
	document []byte, collector *issueCollector,
) (contract.QueryDocument, bool) {
	object, valid := decodeObject(document, "", rootMembers, collector)
	if !valid {
		return contract.QueryDocument{}, false
	}
	input := contract.QueryDocumentInput{}
	assignments := []struct {
		name   string
		target *contract.RawMember
	}{
		{"select", &input.Select}, {"where", &input.Where},
		{"sort", &input.Sort}, {"page", &input.Page},
		{"search", &input.Search}, {"count", &input.Count},
		{"include", &input.Include}, {"aggregate", &input.Aggregate},
	}
	for _, assignment := range assignments {
		raw, present := object[assignment.name]
		if !present {
			continue
		}
		member, err := contract.NewRawMember(raw)
		if err != nil {
			collector.add("", restqueryv1.IssueValueNotCanonical)
			return contract.QueryDocument{}, false
		}
		*assignment.target = member
	}
	result, err := contract.NewQueryDocument(input)
	if err != nil {
		collector.add("", restqueryv1.IssueValueNotCanonical)
		return contract.QueryDocument{}, false
	}
	return result, true
}

func memberValue(member contract.RawMember) (json.RawMessage, bool) {
	raw, present := member.Value()
	return json.RawMessage(raw), present
}

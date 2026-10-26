package v1query

import (
	"cmp"
	"slices"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

type aliasUse struct {
	aliasID string
	path    string
	issue   restqueryv1.IssueCode
}

type aliasSet map[aliasUse]struct{}

func newAliasSet() aliasSet { return make(aliasSet) }

func (set aliasSet) add(
	aliasID string, path string, issue restqueryv1.IssueCode,
) {
	if aliasID == "" {
		return
	}
	set[aliasUse{aliasID: aliasID, path: path, issue: issue}] = struct{}{}
}

func (set aliasSet) list() []aliasUse {
	result := make([]aliasUse, 0, len(set))
	for used := range set {
		result = append(result, used)
	}
	slices.SortFunc(result, compareAliasUses)
	return slices.CompactFunc(result, func(left, right aliasUse) bool {
		return compareAliasUses(left, right) == 0
	})
}

func compareAliasUses(left, right aliasUse) int {
	if order := cmp.Compare(left.aliasID, right.aliasID); order != 0 {
		return order
	}
	if order := cmp.Compare(left.path, right.path); order != 0 {
		return order
	}
	return cmp.Compare(string(left.issue), string(right.issue))
}

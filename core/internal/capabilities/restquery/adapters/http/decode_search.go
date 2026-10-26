package httpquery

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"golang.org/x/text/unicode/norm"
)

var searchMembers = []string{"term"}

func decodeSearch(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (*api.RequestSearch, bool) {
	if !present || rawIsNull(raw) {
		return nil, true
	}
	start := len(collector.issues)
	object, valid := decodeObject(raw, path, searchMembers, collector)
	if !valid {
		return nil, false
	}
	termRaw, termPresent := requireMember(object, path, "term", collector)
	term := ""
	if termPresent {
		term, valid = decodeString(termRaw, rawMemberPath(path, "term"), collector)
	} else {
		valid = false
	}
	if valid {
		valid = validateSearchTerm(term, rawMemberPath(path, "term"), collector)
	}
	if !valid || len(collector.issues) != start {
		return nil, false
	}
	search, err := api.NewRequestSearch(term)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return nil, false
	}
	return &search, true
}

func validateSearchTerm(term, path string, collector *issueCollector) bool {
	normalized := norm.NFC.String(term)
	if strings.ContainsRune(normalized, '\x00') {
		collector.add(path, restqueryv1.IssueNULForbidden)
		return false
	}
	if strings.TrimSpace(normalized) == "" {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return false
	}
	if utf8.RuneCountInString(normalized) > 128 || len(normalized) > 256 {
		collector.addBudget(path, restqueryv1.IssueSearchSizeExceeded)
		return false
	}
	return true
}

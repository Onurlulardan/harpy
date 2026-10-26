package httpquery

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"golang.org/x/text/unicode/norm"
)

func decodeGETFieldName(
	member getMember, collector *issueCollector,
) (string, bool) {
	if member.value == "" {
		collector.add(member.pointer, restqueryv1.IssueNameInvalid)
		return "", false
	}
	if !norm.NFC.IsNormalString(member.value) {
		collector.add(member.pointer, restqueryv1.IssueNFCRequired)
		return "", false
	}
	if member.value == reservedIncludedName {
		collector.add(member.pointer, restqueryv1.IssueNameReserved)
		return "", false
	}
	return member.value, true
}

func decodeGETLiteral(
	member getMember, collector *issueCollector,
) (api.Literal, bool) {
	literal, err := api.NewTextLiteral(member.value)
	if err != nil {
		collector.add(member.pointer, restqueryv1.IssueValueNotCanonical)
		return api.Literal{}, false
	}
	return literal, true
}

func validateGETSearchTerm(
	member getMember, collector *issueCollector,
) (string, bool) {
	normalized := norm.NFC.String(member.value)
	if strings.TrimSpace(normalized) == "" {
		collector.add(member.pointer, restqueryv1.IssueValueNotAllowed)
		return "", false
	}
	if !validateSearchTerm(normalized, member.pointer, collector) {
		return "", false
	}
	return normalized, true
}

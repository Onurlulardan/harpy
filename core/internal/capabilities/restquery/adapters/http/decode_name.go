package httpquery

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"golang.org/x/text/unicode/norm"
)

const reservedIncludedName = "_included"

func decodeName(
	raw []byte, path string, collector *issueCollector,
) (string, bool) {
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return "", false
	}
	if value == "" {
		collector.add(path, restqueryv1.IssueNameInvalid)
		return "", false
	}
	if strings.ContainsRune(value, '\x00') {
		collector.add(path, restqueryv1.IssueNULForbidden)
		return "", false
	}
	if !norm.NFC.IsNormalString(value) {
		collector.add(path, restqueryv1.IssueNFCRequired)
		return "", false
	}
	return value, true
}

func decodeFieldName(
	raw []byte, path string, collector *issueCollector,
) (string, bool) {
	value, valid := decodeName(raw, path, collector)
	if valid && value == reservedIncludedName {
		collector.add(path, restqueryv1.IssueNameReserved)
		return "", false
	}
	return value, valid
}

func validMeasureWireName(value string) bool {
	if len(value) < 1 || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

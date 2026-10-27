package httpquery

import (
	"encoding/json"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeLiteral(
	raw []byte, path string, collector *issueCollector,
) (api.Literal, bool) {
	if rawIsNull(raw) {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return api.Literal{}, false
	}
	literal, err := api.NewJSONLiteral(raw)
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotCanonical)
		return api.Literal{}, false
	}
	return literal, true
}

func decodeLiteralArray(
	raw []byte, path string, collector *issueCollector,
) ([]api.Literal, bool) {
	items, valid := decodeArray(raw, path, collector)
	if !valid {
		return nil, false
	}
	result := make([]api.Literal, 0, len(items))
	for index, item := range items {
		itemPath := rawMemberPath(path, strconv.Itoa(index))
		literal, literalValid := decodeLiteral(json.RawMessage(item), itemPath, collector)
		if !literalValid {
			valid = false
			continue
		}
		result = append(result, literal)
	}
	return result, valid
}

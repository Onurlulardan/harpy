package httpquery

import (
	"encoding/json"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeCount(
	raw json.RawMessage, present bool, path string, collector *issueCollector,
) (api.CountMode, bool) {
	if !present {
		return api.CountModeNone, true
	}
	value, valid := decodeString(raw, path, collector)
	if !valid {
		return 0, false
	}
	mode, found := api.ParseCountMode(value)
	if !found {
		collector.add(path, restqueryv1.IssueValueNotAllowed)
		return 0, false
	}
	return mode, true
}

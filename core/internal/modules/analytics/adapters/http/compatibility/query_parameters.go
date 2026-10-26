package compatibilityhttp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

var legacyFilterPattern = regexp.MustCompile(`^filter\[([A-Za-z][A-Za-z0-9_.-]{0,63})\]$`)

func getParameters(request *http.Request) ([]byte, error) {
	if request == nil || request.URL == nil || len(request.URL.RawQuery) > legacyQueryLimit ||
		!httptransport.BodyAbsent(request) {
		return nil, httptransport.ErrInvalid
	}
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil || len(query) > 104 {
		return nil, httptransport.ErrInvalid
	}
	document := executeDocument{Filters: map[string]json.RawMessage{}}
	for key, values := range query {
		if len(values) != 1 || !validText(values[0], 4096) {
			return nil, httptransport.ErrInvalid
		}
		if match := legacyFilterPattern.FindStringSubmatch(key); match != nil {
			encoded, marshalErr := json.Marshal(values[0])
			if marshalErr != nil {
				return nil, httptransport.ErrInvalid
			}
			document.Filters[match[1]] = encoded
			continue
		}
		if !assignGET(&document, key, values[0]) {
			return nil, httptransport.ErrInvalid
		}
	}
	if len(document.Filters) == 0 {
		document.Filters = nil
	}
	if document.Page == nil {
		document.Page = intAddress(1)
	}
	if document.Limit == nil {
		document.Limit = intAddress(20)
	}
	if !validDocument(document) {
		return nil, httptransport.ErrInvalid
	}
	return canonicalDocument(document)
}

func assignGET(document *executeDocument, key, value string) bool {
	switch key {
	case "page":
		parsed, ok := positiveInteger(value, 1)
		document.Page = intAddress(parsed)
		return ok
	case "limit":
		parsed, ok := positiveInteger(value, 20)
		document.Limit = intAddress(parsed)
		return ok
	case "search":
		document.Search = stringAddress(value)
		return true
	case "sort[field]":
		ensureSort(document).Field = value
		return true
	case "sort[order]":
		ensureSort(document).Order = value
		return true
	default:
		return false
	}
}

func positiveInteger(value string, defaultValue int) (int, bool) {
	if value == "" || value == "0" {
		return defaultValue, true
	}
	parsed, err := strconv.ParseUint(value, 10, 31)
	return int(parsed), err == nil && parsed > 0
}

func ensureSort(document *executeDocument) *sortDocument {
	if document.Sort == nil {
		document.Sort = &sortDocument{}
	}
	return document.Sort
}

func intAddress(value int) *int          { return &value }
func stringAddress(value string) *string { return &value }

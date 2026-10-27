package compatibilityhttp

import (
	"encoding/json"
	"net/http"
	"regexp"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const legacyBodyLimit int64 = 64 << 10
const legacyQueryLimit = 16 << 10

var legacyParameterNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,63}$`)

type sortDocument struct {
	Field string `json:"field"`
	Order string `json:"order,omitempty"`
}

type executeDocument struct {
	Filters map[string]json.RawMessage `json:"filters,omitempty"`
	Params  map[string]json.RawMessage `json:"params,omitempty"`
	Page    *int                       `json:"page,omitempty"`
	Limit   *int                       `json:"limit,omitempty"`
	Sort    *sortDocument              `json:"sort,omitempty"`
	Search  *string                    `json:"search,omitempty"`
}

func postParameters(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	document, err := httptransport.DecodeLimit[executeDocument](writer, request, legacyBodyLimit)
	if err != nil || !validDocument(document) {
		return nil, httptransport.ErrInvalid
	}
	return canonicalDocument(document)
}

func validDocument(document executeDocument) bool {
	if len(document.Filters) > 100 || len(document.Params) > 100 ||
		!validMembers(document.Filters) || !validMembers(document.Params) {
		return false
	}
	if document.Sort != nil &&
		(!validText(document.Sort.Field, 128) || !validText(document.Sort.Order, 16)) {
		return false
	}
	return document.Search == nil || validText(*document.Search, 4096)
}

func validMembers(values map[string]json.RawMessage) bool {
	for name, value := range values {
		if !legacyParameterNamePattern.MatchString(name) || len(value) == 0 || !json.Valid(value) {
			return false
		}
	}
	return true
}

func validText(value string, maximum int) bool {
	return len(value) <= maximum && utf8.ValidString(value) && !containsNUL(value)
}

func canonicalDocument(document executeDocument) ([]byte, error) {
	body, err := json.Marshal(document)
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	defer clear(body)
	canonical, err := canonicaljsonv1.CanonicalizeObject(body, int(legacyBodyLimit))
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	return canonical, nil
}

func containsNUL(value string) bool {
	for _, character := range value {
		if character == 0 {
			return true
		}
	}
	return false
}

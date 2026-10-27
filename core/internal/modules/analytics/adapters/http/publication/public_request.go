package publicationhttp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

const publicBodyLimit int64 = 64 << 10
const publicQueryLimit = 16 << 10

var (
	publicSlugPattern   = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	publicScalarPattern = regexp.MustCompile(`^param\[([A-Za-z][A-Za-z0-9_.-]{0,63})\]$`)
	publicListPattern   = regexp.MustCompile(`^param\[([A-Za-z][A-Za-z0-9_.-]{0,63})\]\[(0|[1-9][0-9]*)\]$`)
)

type publicPOSTDocument struct {
	Params json.RawMessage `json:"params"`
	Page   json.RawMessage `json:"page,omitempty"`
}

func publicIdentity(request *http.Request) ([16]byte, string, publicationapp.ResourceKind, bool) {
	if request == nil {
		return [16]byte{}, "", "", false
	}
	wire := request.PathValue("public_id")
	decoded, err := base64.RawURLEncoding.DecodeString(wire)
	slug := request.PathValue("slug")
	kind := publicationapp.ResourceQuery
	if strings.HasPrefix(request.URL.Path, "/api/published/dashboards/") {
		kind = publicationapp.ResourceDashboard
	}
	if err != nil || len(decoded) != 16 || base64.RawURLEncoding.EncodeToString(decoded) != wire ||
		!publicSlugPattern.MatchString(slug) || len(slug) > 100 {
		return [16]byte{}, "", "", false
	}
	publicID := [16]byte{}
	copy(publicID[:], decoded)
	return publicID, slug, kind, true
}

func publicCapability(request *http.Request) ([]byte, error) {
	values := request.Header.Values("Authorization")
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) != 1 {
		return nil, httptransport.ErrForbidden
	}
	parts := strings.Split(values[0], " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return nil, httptransport.ErrForbidden
	}
	token, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(token) != 32 || base64.RawURLEncoding.EncodeToString(token) != parts[1] {
		clear(token)
		return nil, httptransport.ErrForbidden
	}
	return token, nil
}

func publicHeadersValid(request *http.Request) bool {
	return request != nil && len(request.Header.Values("X-Tenant-ID")) == 0 &&
		len(request.Header.Values("X-Tenant-Slug")) == 0
}

func publicPOSTParameters(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	document, err := httptransport.DecodeLimit[publicPOSTDocument](writer, request, publicBodyLimit)
	if err != nil || len(document.Params) == 0 {
		return nil, httptransport.ErrInvalid
	}
	body, err := json.Marshal(document)
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(body, int(publicBodyLimit))
	clear(body)
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	return canonical, nil
}

func publicGETParameters(request *http.Request) ([]byte, error) {
	if request == nil || request.URL == nil || len(request.URL.RawQuery) > publicQueryLimit ||
		strings.Contains(request.URL.RawQuery, "+") || !httptransport.BodyAbsent(request) {
		return nil, httptransport.ErrInvalid
	}
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil || len(query) > 102 {
		return nil, httptransport.ErrInvalid
	}
	params := map[string]any{}
	lists := map[string]map[int]string{}
	page := map[string]string{}
	for key, values := range query {
		if len(values) != 1 || !utf8.ValidString(key) || !utf8.ValidString(values[0]) ||
			strings.ContainsRune(key, 0) || strings.ContainsRune(values[0], 0) {
			return nil, httptransport.ErrInvalid
		}
		if match := publicScalarPattern.FindStringSubmatch(key); match != nil {
			if _, exists := params[match[1]]; exists || lists[match[1]] != nil {
				return nil, httptransport.ErrInvalid
			}
			params[match[1]] = values[0]
			continue
		}
		if match := publicListPattern.FindStringSubmatch(key); match != nil {
			if _, exists := params[match[1]]; exists {
				return nil, httptransport.ErrInvalid
			}
			index, parseErr := strconv.Atoi(match[2])
			if parseErr != nil || index > 99 {
				return nil, httptransport.ErrInvalid
			}
			if lists[match[1]] == nil {
				lists[match[1]] = map[int]string{}
			}
			lists[match[1]][index] = values[0]
			continue
		}
		if key == "page[size]" || key == "page[after]" {
			page[key[5:len(key)-1]] = values[0]
			continue
		}
		return nil, httptransport.ErrInvalid
	}
	for name, indexed := range lists {
		values := make([]string, len(indexed))
		for index := range values {
			value, exists := indexed[index]
			if !exists {
				return nil, httptransport.ErrInvalid
			}
			values[index] = value
		}
		params[name] = values
	}
	document := map[string]any{"params": params}
	if len(page) != 0 {
		document["page"] = page
	}
	body, err := json.Marshal(document)
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(body, int(publicBodyLimit))
	clear(body)
	if err != nil {
		return nil, httptransport.ErrInvalid
	}
	return canonical, nil
}

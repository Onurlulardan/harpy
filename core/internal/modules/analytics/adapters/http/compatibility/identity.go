package compatibilityhttp

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

var (
	legacyTenantPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,199}$`)
	legacySlugPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]{0,99}$`)
)

func legacyKey(request *http.Request) (compatibilityapp.LegacyKey, error) {
	if request == nil || request.URL == nil ||
		(request.Method != http.MethodGet && request.Method != http.MethodPost) {
		return compatibilityapp.LegacyKey{}, httptransport.ErrInvalid
	}
	tenantIdentity, ok := legacyTenantIdentity(request)
	if !ok {
		return compatibilityapp.LegacyKey{}, httptransport.ErrInvalid
	}
	slug := request.PathValue("slug")
	if !legacySlugPattern.MatchString(slug) || !utf8.ValidString(slug) {
		return compatibilityapp.LegacyKey{}, httptransport.ErrInvalid
	}
	path := request.URL.Path
	if request.URL.RawPath != "" || !legacyPath(path, slug) {
		return compatibilityapp.LegacyKey{}, httptransport.ErrInvalid
	}
	return compatibilityapp.LegacyKey{TenantIdentity: tenantIdentity,
		Method: request.Method, Path: path, Slug: slug}, nil
}

func legacyTenantIdentity(request *http.Request) (string, bool) {
	idValues := request.Header.Values("X-Tenant-ID")
	slugValues := request.Header.Values("X-Tenant-Slug")
	if len(idValues)+len(slugValues) != 1 {
		return "", false
	}
	value := ""
	if len(idValues) == 1 {
		value = idValues[0]
	} else {
		value = slugValues[0]
	}
	return value, legacyTenantPattern.MatchString(value) && strings.TrimSpace(value) == value
}

func legacyPath(path, slug string) bool {
	return path == "/api/openapi/execute/"+slug || path == "/api/analytics/execute/"+slug
}

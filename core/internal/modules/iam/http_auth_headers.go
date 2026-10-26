package iam

import (
	"net/http"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

const (
	authTenantIDHeader   = "X-Tenant-ID"
	authTenantSlugHeader = "X-Tenant-Slug"
)

func exactOrigin(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	values := r.Header.Values("Origin")
	return exactNonempty(values, 512)
}

func exactCSRFHeader(r *http.Request) ([]byte, bool) {
	if r == nil {
		return nil, false
	}
	value, valid := exactNonempty(r.Header.Values("X-CSRF-Token"), 192)
	if !valid {
		return nil, false
	}
	return []byte(value), true
}

func exactBearer(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	value, valid := exactNonempty(
		r.Header.Values("Authorization"), accesstoken.MaximumWireSize+7,
	)
	if !valid || !strings.HasPrefix(value, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(value, "Bearer ")
	return token, token != "" && len(token) <= accesstoken.MaximumWireSize &&
		!strings.ContainsAny(token, " \t\r\n")
}

func exactNonempty(values []string, maximum int) (string, bool) {
	returnValue := ""
	if len(values) == 1 {
		returnValue = values[0]
	}
	return returnValue, returnValue != "" && len(returnValue) <= maximum &&
		strings.TrimSpace(returnValue) == returnValue
}

func tenantHeadersMatch(
	r *http.Request,
	tenantID uuid.UUID,
	tenantSlug string,
) bool {
	return optionalTenantIDMatches(r.Header.Values(authTenantIDHeader), tenantID) &&
		optionalExactMatches(r.Header.Values(authTenantSlugHeader), tenantSlug, 128)
}

func optionalTenantIDMatches(values []string, tenantID uuid.UUID) bool {
	if len(values) == 0 {
		return true
	}
	if len(values) != 1 || tenantID == uuid.Nil {
		return false
	}
	parsed, err := uuid.Parse(values[0])
	return err == nil && parsed == tenantID && parsed.String() == values[0]
}

func optionalExactMatches(values []string, expected string, maximum int) bool {
	if len(values) == 0 {
		return true
	}
	value, valid := exactNonempty(values, maximum)
	return valid && value == expected
}

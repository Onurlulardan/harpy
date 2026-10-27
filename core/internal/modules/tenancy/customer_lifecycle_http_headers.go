package tenancy

import (
	"net/http"
	"strconv"
	"strings"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

func customerLifecycleCredential(r *http.Request) lifecycle.CredentialEnvelope {
	if r == nil {
		return lifecycle.CredentialEnvelope{}
	}
	return lifecycle.CredentialEnvelope{
		AuthorizationValues: append([]string(nil), r.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), r.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), r.Header.Values("X-Tenant-Slug")...),
	}
}

func customerLifecycleTenantID(r *http.Request) (uuid.UUID, bool) {
	if r == nil {
		return uuid.Nil, false
	}
	value := r.PathValue("tenant_id")
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func customerLifecycleOperationID(r *http.Request) (uuid.UUID, bool) {
	if r == nil {
		return uuid.Nil, false
	}
	value := r.PathValue("operation_id")
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

func customerLifecycleIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil || len(r.Header.Values("Idempotency-Key")) != 1 {
		return "", false
	}
	value := r.Header.Get("Idempotency-Key")
	if len(value) < 16 || len(value) > 128 {
		return "", false
	}
	for index, character := range value {
		if !(character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || index > 0 && strings.ContainsRune("._:-", character)) {
			return "", false
		}
	}
	return value, true
}

func customerLifecycleExpectedRevision(r *http.Request) (int64, bool) {
	if r == nil || len(r.Header.Values("If-Match")) != 1 {
		return 0, false
	}
	wire := r.Header.Get("If-Match")
	if len(wire) < 3 || wire[0] != '"' || wire[len(wire)-1] != '"' {
		return 0, false
	}
	value := wire[1 : len(wire)-1]
	revision, err := strconv.ParseInt(value, 10, 64)
	return revision, err == nil && revision > 0 && strconv.FormatInt(revision, 10) == value
}

func customerLifecycleETag(revision int64) string {
	return `"` + strconv.FormatInt(revision, 10) + `"`
}

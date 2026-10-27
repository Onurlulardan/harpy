package tenancy

import (
	"net/http"
	"strconv"
	"strings"
)

func tenantSettingsIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil || len(r.Header.Values("Idempotency-Key")) != 1 {
		return "", false
	}
	value := r.Header.Get("Idempotency-Key")
	if len(value) < 16 || len(value) > 128 {
		return "", false
	}
	for index, character := range value {
		valid := character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || index > 0 && strings.ContainsRune("._:-", character)
		if !valid {
			return "", false
		}
	}
	return value, true
}

func tenantSettingsExpectedRevision(r *http.Request) (int64, bool) {
	if r == nil || len(r.Header.Values("If-Match")) != 1 {
		return 0, false
	}
	value := r.Header.Get("If-Match")
	if len(value) < 3 || value[0] != '"' || value[len(value)-1] != '"' {
		return 0, false
	}
	wire := value[1 : len(value)-1]
	revision, err := strconv.ParseInt(wire, 10, 64)
	return revision, err == nil && revision > 0 && strconv.FormatInt(revision, 10) == wire
}

func tenantSettingsETag(revision int64) string {
	return `"` + strconv.FormatInt(revision, 10) + `"`
}

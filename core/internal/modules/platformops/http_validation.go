package platformops

import (
	"reflect"
	"regexp"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

var (
	idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)
	httpSpanIDPattern  = regexp.MustCompile(`^[0-9a-f]{16,32}$`)
)

func nilHTTPPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

func canonicalHTTPUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func validHTTPTraceID(value string) bool {
	return requestmeta.ValidCorrelationID(value)
}

func validHTTPSpanID(value string) bool { return httpSpanIDPattern.MatchString(value) }

func exactIdempotencyKey(values []string) (string, bool) {
	if len(values) != 1 || !idempotencyPattern.MatchString(values[0]) {
		return "", false
	}
	return values[0], true
}

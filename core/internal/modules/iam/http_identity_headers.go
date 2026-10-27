package iam

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var identityIdempotencyPattern = regexp.MustCompile(
	`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`,
)

func identityPathID(r *http.Request) (uuid.UUID, bool) {
	if r == nil {
		return uuid.Nil, false
	}
	wire := r.PathValue("id")
	id, err := uuid.Parse(wire)
	return id, err == nil && id != uuid.Nil && id.String() == wire
}

func identityIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	value, valid := exactNonempty(r.Header.Values("Idempotency-Key"), 128)
	return value, valid && identityIdempotencyPattern.MatchString(value)
}

func identityExpectedRevision(r *http.Request) (int64, bool) {
	if r == nil {
		return 0, false
	}
	value, valid := exactNonempty(r.Header.Values("If-Match"), 64)
	if !valid || !strings.HasPrefix(value, `"revision-`) ||
		!strings.HasSuffix(value, `"`) {
		return 0, false
	}
	digits := strings.TrimSuffix(strings.TrimPrefix(value, `"revision-`), `"`)
	if digits == "" || digits[0] == '0' {
		return 0, false
	}
	revision, err := strconv.ParseInt(digits, 10, 64)
	return revision, err == nil && revision > 0
}

func identityMutationFromRequest(r *http.Request) (IdentityMutation, bool) {
	id, idValid := identityPathID(r)
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	return IdentityMutation{
		ID: id, ExpectedRevision: revision, IdempotencyKey: key,
	}, idValid && keyValid && revisionValid
}

package iam

import (
	"net/http"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func profilePathResource(r *http.Request) (profileapp.ResourceKind, bool) {
	if r == nil {
		return "", false
	}
	resource := profileapp.ResourceKind(r.PathValue("resource"))
	return resource, resource.Valid()
}

func profilePathUUID(r *http.Request, name string) (uuid.UUID, bool) {
	if r == nil || name == "" {
		return uuid.Nil, false
	}
	return decodeUUID(r.PathValue(name))
}

func profileSelfIdempotencyKey(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	if len(r.Header.Values("Idempotency-Key")) > 0 {
		return identityIdempotencyKey(r)
	}
	key := "profile-self:" + authHTTPCorrelationID(r).String()
	return key, identityIdempotencyPattern.MatchString(key)
}

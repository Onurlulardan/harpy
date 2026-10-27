package iam

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type HTTPErrorSink func(error)

type JWKSHandler struct {
	keys   *accesstoken.PublicKeySet
	report HTTPErrorSink
}

func NewJWKSHandler(
	keys *accesstoken.PublicKeySet,
	report HTTPErrorSink,
) (*JWKSHandler, error) {
	if keys == nil || report == nil {
		return nil, fmt.Errorf("JWKS handler dependencies are required")
	}
	if _, _, _, err := keys.JWKS(); err != nil {
		return nil, fmt.Errorf("JWKS handler key set: %w", err)
	}
	return &JWKSHandler{keys: keys, report: report}, nil
}

func (handler *JWKSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		handler.writeFailure(w, r, httpwire.FailureMethodNotAllowed)
		return
	}
	document, etag, cacheControl, err := handler.keys.JWKS()
	if err != nil {
		handler.report(err)
		handler.writeFailure(w, r, httpwire.FailureServiceUnavailable)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if matchesETag(r.Header.Values("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/jwk-set+json")
	w.Header().Set("Content-Length", strconv.Itoa(len(document)))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := w.Write(document); err != nil {
		handler.report(fmt.Errorf("write JWKS response: %w", err))
	}
}

func (handler *JWKSHandler) writeFailure(
	w http.ResponseWriter,
	r *http.Request,
	failure httpwire.Failure,
) {
	correlationID, found := requestmeta.CorrelationIDFromContext(r.Context())
	if !found {
		correlationID = requestmeta.EntropyFailureCorrelationID
	}
	if err := httpwire.WriteFailure(
		w, correlationID, failure, r.Method == http.MethodHead,
	); err != nil {
		handler.report(err)
	}
}

func matchesETag(values []string, current string) bool {
	for _, value := range values {
		for _, candidate := range strings.Split(value, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "*" || candidate == current ||
				strings.TrimPrefix(candidate, "W/") == current {
				return true
			}
		}
	}
	return false
}

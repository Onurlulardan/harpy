package iam

import (
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *LifecycleHTTP) retired(w http.ResponseWriter, r *http.Request,
	replacement string) {
	if replacement != "" {
		w.Header().Set("Link", "<"+replacement+">; rel=\"successor-version\"")
	}
	w.Header().Set("Cache-Control", "no-store")
	envelope, err := runtimev1.NewErrorEnvelope("CLIENT_UPGRADE_REQUIRED",
		"This V1 endpoint is retired; use the V2 IAM contract",
		authHTTPCorrelationID(r).String(), nil)
	if err == nil {
		err = httpwire.WriteJSON(w, http.StatusGone, authHTTPCorrelationID(r), envelope, false)
	}
	if err != nil {
		handler.report(err)
	}
}

package iam

import (
	"errors"
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *SystemTransferHTTP) systemTransferUpgradeRequired(w http.ResponseWriter,
	r *http.Request, successor string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Link", "<"+successor+">; rel=\"successor-version\"")
	w.Header().Set("Deprecation", "true")
	envelope, err := runtimev1.NewErrorEnvelope("CLIENT_UPGRADE_REQUIRED",
		"This V1 IAM transfer endpoint is retired; use the asynchronous V2 contract",
		systemTransferCorrelationID(r).String(), nil)
	if err == nil {
		err = httpwire.WriteJSON(w, http.StatusUpgradeRequired,
			systemTransferCorrelationID(r), envelope, false)
	}
	if err != nil && handler != nil && handler.report != nil {
		handler.report(errors.New("IAM transfer V1 retirement response failed"))
	}
}

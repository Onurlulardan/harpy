package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *LifecycleHTTP) writeJSON(w http.ResponseWriter, r *http.Request,
	status int, document any) {
	if err := httpwire.WriteJSON(w, status, authHTTPCorrelationID(r), document, false); err != nil {
		handler.report(err)
	}
}

func (handler *LifecycleHTTP) noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

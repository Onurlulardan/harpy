package management

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func (handler *Handler) writeVersion(
	w http.ResponseWriter,
	r *http.Request,
	correlationID requestmeta.CorrelationID,
) {
	if err := httpwire.WriteJSON(
		w,
		http.StatusOK,
		correlationID,
		handler.info.Version(),
		r.Method == http.MethodHead,
	); err != nil {
		handler.report(err)
	}
}

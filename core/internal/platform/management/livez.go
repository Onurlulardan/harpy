package management

import (
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func (handler *Handler) writeLivez(
	w http.ResponseWriter,
	r *http.Request,
	correlationID requestmeta.CorrelationID,
) {
	payload, err := runtimev1.NewLivez(handler.info.Service())
	if err != nil {
		handler.report(err)
		handler.writeFailure(w, r, correlationID, httpwire.FailureInternal)
		return
	}
	if err := httpwire.WriteJSON(w, http.StatusOK, correlationID, payload, r.Method == http.MethodHead); err != nil {
		handler.report(err)
	}
}

package management

import (
	"net/http"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func (handler *Handler) writeReadyz(
	w http.ResponseWriter,
	r *http.Request,
	correlationID requestmeta.CorrelationID,
) {
	payload, err := handler.monitor.Snapshot()
	if err != nil {
		handler.report(err)
		handler.writeFailure(w, r, correlationID, httpwire.FailureInternal)
		return
	}
	status := http.StatusServiceUnavailable
	if payload.Status == runtimev1.ReadinessStatusReady {
		status = http.StatusOK
	} else {
		w.Header().Set("Retry-After", "1")
	}
	if err := httpwire.WriteJSON(w, status, correlationID, payload, r.Method == http.MethodHead); err != nil {
		handler.report(err)
	}
}

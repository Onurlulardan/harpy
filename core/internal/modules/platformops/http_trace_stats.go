package platformops

import "net/http"

func (handler *HTTPHandler) getTraceStats(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.traceEnvironment(r)
	if err == nil {
		stats, readErr := handler.traces.ReadTraceStats(r.Context(), environment.EnvironmentID)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, traceStatsAPI(stats))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

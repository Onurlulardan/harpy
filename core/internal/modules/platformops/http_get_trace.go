package platformops

import "net/http"

func (handler *HTTPHandler) getTrace(w http.ResponseWriter, r *http.Request) {
	correlationID := r.PathValue("correlation_id")
	if !validHTTPTraceID(correlationID) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, err := handler.traceEnvironment(r)
	if err == nil {
		spans, readErr := handler.traces.GetTrace(
			r.Context(), environment.EnvironmentID, correlationID)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, traceDetailAPI(spans))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

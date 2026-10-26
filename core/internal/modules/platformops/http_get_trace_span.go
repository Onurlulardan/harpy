package platformops

import "net/http"

func (handler *HTTPHandler) getTraceSpan(w http.ResponseWriter, r *http.Request) {
	spanID := r.PathValue("span_id")
	if !validHTTPSpanID(spanID) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, err := handler.traceEnvironment(r)
	if err == nil {
		span, readErr := handler.traces.GetTraceSpan(
			r.Context(), environment.EnvironmentID, spanID)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, traceSpanAPI(span))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

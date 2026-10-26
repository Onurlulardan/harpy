package platformops

import "net/http"

func (handler *HTTPHandler) listTraces(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.traceEnvironment(r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	filter, err := traceFilterFromQuery(r)
	if err == nil {
		result, readErr := handler.traces.ListTraces(r.Context(), environment.EnvironmentID, filter)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, tracePageAPI(result))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

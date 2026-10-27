package platformops

import "net/http"

func (handler *HTTPHandler) getAlertEvent(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("alert_event_id")
	if !canonicalHTTPUUID(eventID) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, _, err := handler.alertEnvironment(
		r, "platform.operations.alerts.read")
	if err == nil {
		event, readErr := handler.alerts.GetAlertEvent(
			r.Context(), environment.EnvironmentID, eventID)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, alertEventAPI(event))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

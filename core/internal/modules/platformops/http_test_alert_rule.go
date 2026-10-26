package platformops

import "net/http"

func (handler *HTTPHandler) testAlertRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("alert_rule_id")
	if !canonicalHTTPUUID(id) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, access, err := handler.alertEnvironment(
		r, "platform.operations.alerts.manage")
	if err == nil {
		event, testErr := handler.alerts.RecordAlertTest(
			r.Context(), environment.EnvironmentID, id, access.ActorID)
		err = testErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, alertEventAPI(event))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

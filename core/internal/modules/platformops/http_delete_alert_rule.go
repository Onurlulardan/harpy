package platformops

import "net/http"

func (handler *HTTPHandler) deleteAlertRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("alert_rule_id")
	if !canonicalHTTPUUID(id) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, _, err := handler.alertEnvironment(r, "platform.operations.alerts.manage")
	if err == nil {
		err = handler.alerts.DeleteAlertRule(r.Context(), environment.EnvironmentID, id)
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, map[string]bool{"deleted": true})
			return
		}
	}
	handler.writeFailure(w, r, err)
}

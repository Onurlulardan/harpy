package platformops

import "net/http"

func (handler *HTTPHandler) getAlertRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.PathValue("alert_rule_id")
	if !canonicalHTTPUUID(ruleID) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, _, err := handler.alertEnvironment(
		r, "platform.operations.alerts.read")
	if err == nil {
		rule, readErr := handler.alerts.GetAlertRule(
			r.Context(), environment.EnvironmentID, ruleID)
		err = readErr
		if err == nil {
			handler.writeSuccess(w, r, http.StatusOK, alertRuleAPI(rule))
			return
		}
	}
	handler.writeFailure(w, r, err)
}

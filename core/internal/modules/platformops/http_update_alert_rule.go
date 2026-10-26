package platformops

import "net/http"

func (handler *HTTPHandler) updateAlertRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("alert_rule_id")
	if !canonicalHTTPUUID(id) {
		handler.writeFailure(w, r, errHTTPRequestInvalid)
		return
	}
	environment, _, err := handler.alertEnvironment(r, "platform.operations.alerts.manage")
	if err == nil {
		input, inputErr := decodeAlertRule(w, r, environment.EnvironmentID)
		err = inputErr
		if err == nil {
			rule, updateErr := handler.alerts.UpdateAlertRule(r.Context(), id, input)
			err = updateErr
			if err == nil {
				handler.writeSuccess(w, r, http.StatusOK, alertRuleAPI(rule))
				return
			}
		}
	}
	handler.writeFailure(w, r, err)
}

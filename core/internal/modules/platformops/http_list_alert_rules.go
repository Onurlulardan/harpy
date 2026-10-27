package platformops

import "net/http"

func (handler *HTTPHandler) listAlertRules(w http.ResponseWriter, r *http.Request) {
	environment, _, err := handler.alertEnvironment(r, "platform.operations.alerts.read")
	if err == nil {
		rules, readErr := handler.alerts.ListAlertRules(r.Context(), environment.EnvironmentID)
		err = readErr
		if err == nil {
			views := make([]alertRuleView, len(rules))
			for index, rule := range rules {
				views[index] = alertRuleAPI(rule)
			}
			handler.writeSuccess(w, r, http.StatusOK, views)
			return
		}
	}
	handler.writeFailure(w, r, err)
}

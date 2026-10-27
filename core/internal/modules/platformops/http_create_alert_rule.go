package platformops

import "net/http"

func (handler *HTTPHandler) createAlertRule(w http.ResponseWriter, r *http.Request) {
	environment, _, err := handler.alertEnvironment(r, "platform.operations.alerts.manage")
	if err == nil {
		var inputErr error
		input, inputErr := decodeAlertRule(w, r, environment.EnvironmentID)
		err = inputErr
		if err == nil {
			rule, createErr := handler.alerts.CreateAlertRule(r.Context(), input)
			err = createErr
			if err == nil {
				handler.writeSuccess(w, r, http.StatusCreated, alertRuleAPI(rule))
				return
			}
		}
	}
	handler.writeFailure(w, r, err)
}

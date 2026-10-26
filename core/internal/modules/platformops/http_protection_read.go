package platformops

import "net/http"

func (handler *HTTPHandler) getProtection(w http.ResponseWriter, r *http.Request) {
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.recovery.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	protection, err := handler.protection.GetProtection(
		r.Context(), environment.EnvironmentID)
	if err != nil || protection.EnvironmentID != environment.EnvironmentID {
		if err == nil {
			err = ErrHTTPUnavailable
		}
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, protection)
}

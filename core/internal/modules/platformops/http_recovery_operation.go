package platformops

import "net/http"

func (handler *HTTPHandler) getRecoveryOperation(w http.ResponseWriter, r *http.Request) {
	operationID := r.PathValue("operation_id")
	environment, err := handler.onPremiseEnvironmentFor(r)
	if err == nil && !canonicalHTTPUUID(operationID) {
		err = errHTTPRequestInvalid
	}
	if err == nil {
		_, err = handler.accessFor(r, environment, "platform.recovery.read", "")
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	operation, err := handler.protection.GetControlOperation(r.Context(), operationID)
	if err != nil || operation.EnvironmentID != environment.EnvironmentID {
		if err == nil {
			err = ErrHTTPForbidden
		}
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, representRecoveryOperation(operation))
}

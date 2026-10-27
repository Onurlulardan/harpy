package notificationhttp

import "net/http"

func (handler *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionPreferencesRead, OperationPreferencesGet)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	values, err := handler.application.GetPreferences(r.Context(), access)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, viewPreferences(values))
}

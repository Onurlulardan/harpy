package notificationhttp

import "net/http"

func (handler *Handler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionPreferencesEdit, OperationPreferencesEdit)
	var document preferencesDocument
	if err == nil {
		err = decodeJSON(w, r, &document)
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	values, err := handler.application.UpdatePreferences(r.Context(), document.input(access))
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, viewPreferences(values))
}

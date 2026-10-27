package notificationhttp

import "net/http"

func (handler *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionInboxUpdate, OperationInboxReadAll)
	if err == nil {
		err = requireEmptyBody(r)
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	count, err := handler.application.MarkAllRead(r.Context(), access)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, struct {
		Updated int64 `json:"updated"`
	}{Updated: count})
}

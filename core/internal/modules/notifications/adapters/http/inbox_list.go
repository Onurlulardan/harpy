package notificationhttp

import "net/http"

func (handler *Handler) list(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionInboxRead, OperationInboxList)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	filter, err := decodeInboxQuery(r.URL.RawQuery)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	page, err := handler.application.ListMessages(r.Context(), ListInput{Access: access, Filter: filter})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	view, err := viewMessagePage(page)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, view)
}

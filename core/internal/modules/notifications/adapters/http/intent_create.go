package notificationhttp

import "net/http"

func (handler *Handler) createIntent(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionSend, OperationIntentCreate)
	var document intentDocument
	if err == nil {
		err = decodeJSON(w, r, &document)
	}
	var request IntentInput
	if err == nil {
		request.Access = access
		request.Request, err = document.request()
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.application.RecordIntent(r.Context(), request)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	status := http.StatusCreated
	if result.Replay {
		status = http.StatusOK
	}
	handler.writeSuccess(w, r, status, struct {
		MessageID     string `json:"message_id"`
		InboxRevision int64  `json:"inbox_revision"`
		Replay        bool   `json:"replay"`
		Outcome       string `json:"outcome"`
	}{result.MessageID.String(), result.InboxRevision, result.Replay, result.Outcome})
}

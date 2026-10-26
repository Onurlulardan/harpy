package notificationhttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (handler *Handler) get(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionInboxRead, OperationInboxGet)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	id, err := domain.ParseID(r.PathValue("notification_id"))
	if err != nil {
		handler.writeFailure(w, r, ErrInvalid)
		return
	}
	message, err := handler.application.GetMessage(r.Context(), MessageInput{Access: access, ID: id})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, viewMessage(message))
}

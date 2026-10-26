package notificationhttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (handler *Handler) dismiss(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionInboxDelete, OperationInboxDismiss)
	if err == nil {
		err = requireEmptyBody(r)
	}
	id, parseErr := domain.ParseID(r.PathValue("notification_id"))
	if err == nil && parseErr != nil {
		err = ErrInvalid
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	message, err := handler.application.DismissMessage(r.Context(), MessageInput{Access: access, ID: id})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, viewMessage(message))
}

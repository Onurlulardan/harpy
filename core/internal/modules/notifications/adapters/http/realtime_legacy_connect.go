package notificationhttp

import (
	"net/http"

	"github.com/google/uuid"
)

func (handler *Handler) connectLegacy(w http.ResponseWriter, r *http.Request) {
	if err := validateSocketHandshake(r); err != nil {
		handler.writeFailure(w, r, ErrInvalid)
		return
	}
	access, err := handler.accessFor(r, PermissionInboxRead, OperationLegacyConnect)
	userID, parseErr := uuid.Parse(r.PathValue("user_id"))
	if err == nil && (parseErr != nil || userID != access.PrincipalID()) {
		err = ErrForbidden
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	identity, err := handler.application.ConsumeRealtimeTicket(r.Context(), RealtimeInput{
		Access: access, OfferedProtocols: offeredProtocols(r), Legacy: true,
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if err = handler.realtime.Serve(w, r, identity); err != nil {
		handler.report(errRealtimeUnavailable)
	}
}

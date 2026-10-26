package notificationhttp

import "net/http"

func (handler *Handler) connect(w http.ResponseWriter, r *http.Request) {
	if err := validateSocketHandshake(r); err != nil {
		handler.writeFailure(w, r, ErrInvalid)
		return
	}
	access, err := handler.accessFor(r, PermissionInboxRead, OperationSocketConnect)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	identity, err := handler.application.ConsumeRealtimeTicket(r.Context(), RealtimeInput{
		Access: access, OfferedProtocols: offeredProtocols(r),
	})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	if err = handler.realtime.Serve(w, r, identity); err != nil {
		handler.report(errRealtimeUnavailable)
	}
}

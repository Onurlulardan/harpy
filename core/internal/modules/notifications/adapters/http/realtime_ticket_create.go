package notificationhttp

import "net/http"

func (handler *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	access, err := handler.accessFor(r, PermissionInboxRead, OperationTicketCreate)
	if err == nil {
		err = requireEmptyBody(r)
	}
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	result, err := handler.application.CreateRealtimeTicket(r.Context(), access)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusCreated, struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}{Token: result.Token, ExpiresAt: result.ExpiresAt.Format("2006-01-02T15:04:05.999999999Z07:00")})
}

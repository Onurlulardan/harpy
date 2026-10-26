package notificationhttp

import "net/http"

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		{http.MethodGet, "/api/notifications", http.HandlerFunc(handler.list)},
		{http.MethodGet, "/api/notifications/{notification_id}", http.HandlerFunc(handler.get)},
		{http.MethodPut, "/api/notifications/{notification_id}/read", http.HandlerFunc(handler.markRead)},
		{http.MethodPut, "/api/notifications/mark-all-as-read", http.HandlerFunc(handler.markAllRead)},
		{http.MethodDelete, "/api/notifications/{notification_id}", http.HandlerFunc(handler.dismiss)},
		{http.MethodGet, "/api/notification-preferences", http.HandlerFunc(handler.getPreferences)},
		{http.MethodPut, "/api/notification-preferences", http.HandlerFunc(handler.updatePreferences)},
		{http.MethodPost, "/api/notification-intents", http.HandlerFunc(handler.createIntent)},
		{http.MethodPost, "/api/notification-realtime-tickets", http.HandlerFunc(handler.createTicket)},
		{http.MethodGet, "/ws/notifications", http.HandlerFunc(handler.connect)},
		{http.MethodGet, "/ws/notifications/{user_id}", http.HandlerFunc(handler.connectLegacy)},
	}
}

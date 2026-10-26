package audithttp

import "net/http"

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		{http.MethodGet, "/api/changelog", http.HandlerFunc(handler.list)},
		{http.MethodGet, "/api/changelog/stats", http.HandlerFunc(handler.stats)},
		{http.MethodGet, "/api/changelog/{id}", http.HandlerFunc(handler.get)},
		{http.MethodGet, "/api/changelog/history/{entity_type}/{entity_id}",
			http.HandlerFunc(handler.entityHistory)},
		{http.MethodGet, "/api/changelog/forms/{form_slug}/records/{record_id}/history",
			http.HandlerFunc(handler.formRecordHistory)},
	}
}

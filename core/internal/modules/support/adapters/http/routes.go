package supporthttp

import "net/http"

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		{http.MethodPost, "/api/contact", http.HandlerFunc(handler.submitContact)},
		{http.MethodGet, "/api/public/booking-url", http.HandlerFunc(handler.getBooking)},
		{http.MethodGet, "/api/admin/contact-messages", http.HandlerFunc(handler.listContacts)},
		{http.MethodGet, "/api/admin/contact-messages/stats", http.HandlerFunc(handler.contactStats)},
		{http.MethodGet, "/api/admin/contact-messages/{id}", http.HandlerFunc(handler.getContact)},
		{http.MethodPut, "/api/admin/contact-messages/{id}", http.HandlerFunc(handler.updateContact)},
		{http.MethodPost, "/api/bug-reports", http.HandlerFunc(handler.createBug)},
		{http.MethodGet, "/api/bug-reports/mine", http.HandlerFunc(handler.listMyBugs)},
		{http.MethodGet, "/api/bug-reports/{id}", http.HandlerFunc(handler.getMyBug)},
		{http.MethodGet, "/api/bug-reports/admin", http.HandlerFunc(handler.listTenantBugs)},
		{http.MethodGet, "/api/bug-reports/admin/stats", http.HandlerFunc(handler.tenantBugStats)},
		{http.MethodGet, "/api/bug-reports/admin/{id}", http.HandlerFunc(handler.getTenantBug)},
		{http.MethodPut, "/api/bug-reports/admin/{id}", http.HandlerFunc(handler.updateTenantBug)},
		{http.MethodGet, "/api/admin/bug-reports", http.HandlerFunc(handler.listPlatformBugs)},
		{http.MethodGet, "/api/admin/bug-reports/stats", http.HandlerFunc(handler.platformBugStats)},
		{http.MethodGet, "/api/admin/bug-reports/{id}", http.HandlerFunc(handler.getPlatformBug)},
		{http.MethodPut, "/api/admin/bug-reports/{id}", http.HandlerFunc(handler.updatePlatformBug)},
	}
}

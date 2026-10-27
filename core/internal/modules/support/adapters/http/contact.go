package supporthttp

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type contactCreateDocument struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Company       string `json:"company"`
	Phone         string `json:"phone"`
	EmployeeCount string `json:"employee_count"`
	Message       string `json:"message"`
	ContactType   string `json:"contact_type"`
	Source        string `json:"source"`
	Honeypot      string `json:"_honeypot"`
}

type contactUpdateDocument struct {
	Status           *domain.ContactStatus `json:"status"`
	Notes            *string               `json:"notes"`
	ExpectedRevision int64                 `json:"expected_revision"`
}

func (handler *Handler) submitContact(response http.ResponseWriter, request *http.Request) {
	var document contactCreateDocument
	canonical, err := decodeObject(response, request, &document)
	if err != nil {
		handler.writeFailure(response, request, err)
		return
	}
	key, digest, err := commandEvidence(request, canonical)
	if err != nil {
		handler.writeFailure(response, request, err)
		return
	}
	rateKey, err := handler.rateKeys.Resolve(request)
	if err == nil {
		err = handler.applications.Contact.SubmitContact(request.Context(), domain.ContactInput{
			Name: document.Name, Email: document.Email, Company: document.Company,
			Phone: document.Phone, EmployeeCount: document.EmployeeCount,
			Message: document.Message, ContactType: document.ContactType,
			Source: document.Source, Honeypot: document.Honeypot,
		}, rateKey, key, digest)
	}
	if err != nil {
		handler.writeFailure(response, request, err)
		return
	}
	handler.writeSuccess(response, request, http.StatusOK, map[string]any{
		"success": true, "message": "Mesajınız başarıyla alındı.",
	})
}

func (handler *Handler) getBooking(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		handler.writeFailure(response, request, application.ErrInvalid)
		return
	}
	data := map[string]string{"url": handler.booking.URL, "provider": handler.booking.Provider}
	if handler.booking.URL == "" {
		data["message"] = "Booking system not configured"
	}
	handler.writeSuccess(response, request, http.StatusOK, data)
}

func (handler *Handler) listContacts(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformContactRead, application.OperationContactList)
	if err != nil {
		handler.writeFailure(response, request, err)
		return
	}
	filter, err := contactFilter(request)
	if err == nil {
		var page domain.ContactPage
		page, err = handler.applications.Contact.ListContacts(request.Context(), access, filter)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, contactPageView(page))
			return
		}
	}
	handler.writeFailure(response, request, err)
}

func (handler *Handler) contactStats(response http.ResponseWriter, request *http.Request) {
	access, err := handler.accessFor(request, application.ScopePlatform,
		application.PermissionPlatformContactRead, application.OperationContactStats)
	if err == nil && len(request.URL.Query()) == 0 {
		var stats domain.ContactStats
		stats, err = handler.applications.Contact.ReadContactStats(request.Context(), access)
		if err == nil {
			handler.writeSuccess(response, request, http.StatusOK, contactStatsView(stats))
			return
		}
	} else if err == nil {
		err = application.ErrInvalid
	}
	handler.writeFailure(response, request, err)
}
